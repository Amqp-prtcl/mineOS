# SITE:

### **GET**  `/`
> redirects to `/servers`

### **GET**  `/home`
> redirects to `/servers`

### **GET**  `/servers`
> servers html page

### **GET**  `/servers/{serverID}`
> server dashboard (html) page

### **POST** `/servers/{serverID}/start`
> starts server (server must be closed)

### **POST** `/servers/{serverID}/stop`
> stops server (server must be running or Starting)

### **GET** `/assets/{path-to-file}`
> returns content of file (if it exists)

# API:

## Versions: 

### **GET** `/api/versions` 
> returns a json list of available server types (srvType)

example:
```
[
    "vanilla",
    "paper"
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
        "id": "6952360055545620384",
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
    "id": "6952360055534518272",
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
> TODO: create a new server not ready

exemple:
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

# WEBSOCKETS

### **ANY**  `/servers/ws`
> opens a websocket connection to server for servers state changes events

### **ANY**  `/servers/{serverID}/ws`
> opens a websocket connection to server for state changes and minecraft console log events