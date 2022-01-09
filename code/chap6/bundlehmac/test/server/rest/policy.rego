package test.server.rest

default allowed = false

allowed {
    name := input.name
    data.users.users[name].role == "manager"  
}