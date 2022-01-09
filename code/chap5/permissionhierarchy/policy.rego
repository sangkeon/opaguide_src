package example.permissionhierarchy

default allowed = false

allowed {
    input.user.role == "ADMIN"
}

allowed {
    input.user.role == "USER"

    input.user.target_user_id != ""
    
    input.user.target_user_id == input.user.id
}

allowed {
    input.user.role == "OPERATOR"    

    permission := data.operator_permission[input.user.id]
    required_permission := data.api[input.api.url][input.api.method]
    
    satisfied_permission := {p | permissionmatch(permission[_], required_permission[p], ".")}

    count(required_permission) > 0
    count(satisfied_permission) == count(required_permission)
}

permissionmatch(permission, req_permission, delim) = true {
    permission == req_permission
} else = result {
    result := startswith(req_permission, concat("", [permission, delim]))
}

allowed {
    required_permission := data.api[input.api.url][input.api.method]

    count(required_permission) == 0
}