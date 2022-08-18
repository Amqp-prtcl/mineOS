package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// it is safe for concurrent use
type Config struct {
	Config   map[string]interface{}
	Filepath string
	mu       sync.RWMutex
}

// Default values are optionals (set to nil or use an empty map to skip it); LoadConfigFile will use the map to make up
// for all value present in Default but not in file.
//
// Config package only accepts strings; booleans; float64; and structs or arrays containing them.
// ints; units; and float32 are saved as float64
//
// Warning: LoadConfigFile only does shallow copies of values in default (take care about race conditions)
func LoadConfigFile(filepath string, defaults map[string]interface{}) (*Config, error) {
	var config = &Config{
		Config:   map[string]interface{}{},
		Filepath: filepath,
		mu:       sync.RWMutex{},
	}
	defer func() {
		for k, v := range defaults { // range already check for nil maps
			if _, ok := config.Config[k]; !ok {
				config.Config[k] = v
			}
		}
	}()
	f, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	defer f.Close()
	return config, json.NewDecoder(f).Decode(&config.Config)
}

func (c *Config) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	v, ok := c.Config[key]
	c.mu.RUnlock()
	return v, ok
}

func (c *Config) Put(key string, val interface{}) {
	c.mu.Lock()
	c.Config[key] = val
	c.mu.Unlock()
}

// SyncWithDefaults will use the map to make up for all value present in default but not in file.
//
// Warning: SyncWithDefaults only does shallow copies of values in default (take care about race conditions)
func (c *Config) SyncWithDefaults(defaults map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range defaults {
		if _, ok := c.Config[k]; !ok {
			c.Config[k] = v
		}
	}
}

// Warning: GetCopyOfConfig only does shallow copies of values in default (take care about race conditions)
func (c *Config) GetCopyOfConfig() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	var m = map[string]interface{}{}
	for k, v := range c.Config {
		m[k] = v
	}
	return m
}

func (c *Config) SaveFile() error {
	f, err := os.Create(c.Filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.NewEncoder(f).Encode(c.Config)
}

type Getable interface {
	~string | ~float64 | []interface{} | map[string]interface{}
}

func Get[T Getable](config *Config, key string) (T, bool) {
	var ret T
	v, ok := config.Get(key)
	if !ok {
		return ret, false
	}

	ret, ok = v.(T)
	return ret, ok
}

// same as Get but gives more detail about what failed
func GetErr[T Getable](config *Config, key string) (T, error) {
	var ret T
	v, ok := config.Get(key)
	if !ok {
		return ret, ErrKeyNotFound
	}
	ret, ok = v.(T)
	if !ok {
		return ret, fmt.Errorf("config file Get: failed to cast value (wanted type: %T but got type: %T)", ret, v)
	}
	return ret, nil
}
