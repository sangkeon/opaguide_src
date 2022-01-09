package example.urlhierarchy

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

    method := input.api.method
    url := longest_match_api(data.api, input.api.url, method)

    permission := data.operator_permission[input.user.id]
    required_permssion := data.api[url][method]
    
    satisfied_permission := {p | permissionmatch(permission[_], required_permssion[p], ".")}

    count(required_permssion) > 0
    count(satisfied_permission) == count(required_permssion)
}

allowed {
    method := input.api.method
    url := longest_match_api(data.api, input.api.url, method)
 
    url

    required_permssion := data.api[url][method]

    count(required_permssion) == 0
}

permissionmatch(permission, req_permission, delim) = true {
    permission == req_permission
} else = result {
    result := startswith(req_permission, concat("", [permission, delim]))
}