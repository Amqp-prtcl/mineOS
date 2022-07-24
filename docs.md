# SITE:

### **GET** `/login`
> serves login html page

### **POST** `/login`
> returns auth jwt cookie in exchange of valid credentials

example:
```
username=myUsername&password=MySecretPassword
```

NOTE: Notice that at least for now the password is sent in plain text and MineOs does not uses https so a password unique for MineOs is recommended as any proxy or man-in-the-middle attack will be able to access it

### **ANY** `/logout`
> overwrites auth jwt cookie

### **GET**  `/`
>  redirects to `/home`

### **GET**  `/home`
> serves home html page

### **GET**  `/servers`
> serves servers html page

### **GET**  `/servers/{serverID}`
> server dashboard (html) page

### **POST** `/servers/{serverID}/start`
> starts server (server must be closed)

### **POST** `/servers/{serverID}/stop`
> stops server (server must be running or Starting)

### **POST** `/servers/{serverID}/zip`
> will create compress the server into a zip archive before returning the downloadID

example:
```
{
    "download-id":"5689032658932",
}
```

### **GET** `/assets/{path-to-file}`
> returns content of file (if it exists)

### **GET** `/downloads/{downloadID}`
> returns the content of the file

### **GET** `/downloads/{downloadID}/info`
> returns info about the file

example:
```
{
    "name:"server at.zip",
    "size":4239503,
    "sha526":"",
    "expiration-stamp":"79503890532"
}
```

<br>

# API:

## Types:

> all endpoints either send or receive valid json values and require to be authenticated.

### Common types:
- `ID`:
> IDs are string encoded snowflakes (int64)

- `srvType`:
> srvTypes are a string representing a different minecraft version type (ex: "VANILLA"; "PAPER"); you can get a list of all available types at endpoint `GET /api/versions`

- `versionID`:
> versionIDs are a string representing a minecraft version (ex: "1.8.8"; "1.19"); they are usually coupled with srvTypes to represent a server's minecraft characteristics

## General:

### **GET** `/api/epoch`
> return internal epoch used for id generation and interpretation

example:
```
{
    "epoch": "2022-07-14T09:52:24.06398+02:00"
}
```

## Versions: 

### **GET** `/api/versions` 
> returns a json list of available server types (srvType)

example:
```
[
    "VANILLA",
    "PAPER"
]
```

### **GET** `/api/versions/{srvType}` 
> returns a json list of all possible versions IDs for selected server type

example:
```
[
    "1.8.8",
    "1.9",
    "1.12.2",
    "1.19"
]
```
---

## Servers:

### **GET**  `/api/servers`
> returns a list of available servers

example:
```
[
    {
        "id": "6952360055534518272",
        "name": "Example #1",
        "server-type": "VANILLA",
        "version-id": "1.19",
        "state": "RUNNING"
    },
    {
        "id": "6952705593643630592",
        "name": "Example #2",
        "server-type": "PAPER",
        "version-id": "1.8.8",
        "state": "STOPPING"
    }
]
```

### **GET**  `/api/servers/{serverID}`
> returns info about the server

example:
```
{
    "id": "6952705687906418688",
    "name": "Example #1",
    "emails": [
        "fist.example@mail.com",
        "mail.second@mail.com",
    ],
    "server-type": "VANILLA",
    "version-id": "1.19",
    "state": "RUNNING"
}
```

### **POST** `/api/servers/{serverID}/emails`
> send list of emails to be added to server

example:
```
[
    "address.to@mail.com",
    "be.added@mail.com
]
```

### **POST** `/api/servers/new` 
> creates a new server. Name field must not be empty and it is recommended that it is unique in order to avoid some confusion.

example:
```
{
    "name": "New Server", 
    "emails": [
        "incre.dible@mail.com",
        "magni.ficent@mail.com"
    ],
    "server-type": "PAPER",
    "version-id": "1.18"
}
```
if success, returns:
```
{
    "id":"6953766549635203072"
}
```

<br>

# WEBSOCKETS

## Events structure:
| field id | value type |
| - | - |
| event | event Type (String) |
| data | json encoded data of that event type (string) |

example: 
```
{
    "event": "state-update",
    "data": "{\"server-id\": \"6952705532792668160\",
    \"state\": "CLOSED\"}"
}
```

### Event Types:
- `state-update`:

example:
```
{
    "server-id": "6952705532792668160",
    "state": "CLOSED"
}
```

- `log-update`:

example:
```
{
    "server-id": "6953253318667796480",
    "log": "[16:18:48] [Server thread/INFO]: Starting minecraft server version 1.19"
}
```
- `cmd-input`:

example:
```
{
    "server-id": "6953256559354839040"
    "command": "list"
}
```

### **ANY**  `/servers/ws`

> opens a websocket connection to server for servers state changes events

this connection will only send `state-update` events and should not be written to.

### **ANY**  `/servers/{serverID}/ws`
> opens a websocket connection to server for state changes and minecraft console log events

this connection will send `state-update` and `log-update` and can only receive `cmd-input` json objects (if IDs do not match, the event is discarded)

<br>

# TODO LIST
- [X] add offline mode
- [X] sanitize upon profile generation error (if generation fails on later stage (agreeing to EULA), dead folder will remain on disk -> Must remove it)
- [X] option to zip and download backup
- [X] add caching system for versions
- [ ] add JSON config file for each server type (ex: manifest URL, etc)
- [ ] add way of clearing cache (if possible per serverType)
- [ ] auto updates -> auto check and update with the press of a button (just need to replace .jar file) (only present for modded versions)
- [ ] add Bukkit and Spigot support (buildTools.jar)