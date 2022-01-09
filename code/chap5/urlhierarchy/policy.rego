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
    url := longestmatchapi(input.api.url, method)

    permission := data.operator_permission[input.user.id]
    required_permission := data.api[url][method]
    
    satisfied_permission := {p | permissionmatch(permission[_], required_permission[p], ".")}

    count(required_permission) > 0
    count(satisfied_permission) == count(required_permission)
}

allowed {
    method := input.api.method
    url := longestmatchapi(input.api.url, method)
 
    url

    required_permission := data.api[url][method]

    count(required_permission) == 0
}

permissionmatch(permission, req_permission, delim) = true {
    permission == req_permission
} else = result {
    result := startswith(req_permission, concat("", [permission, delim]))
}

longestmatchapi(url, method) = apimatch {
    urlslice := split(url, "/")
    ns := numbers.range(count(urlslice),1)
    apimatch := [api|api := concat("/", array.slice(urlslice, 0, ns[_])); data.api[api][method]; api != "" ][0]
}

# longestmatchapi(url, method) = "" {
#   contains(url, "/") == false
# } else = url {
#   count(data.api[url][method]) > 0
# } else = parent_match {
#     path := split("/",url)
#     num := count(path)
#     last := path[num - 1]

#     parent_api :=  trim_suffix(url, concat("", ["/", last]) )

#     parent_match := longestmatchapi2(parent_api, method)
# }

# longestmatchapi2(url, method) = "" {
#   contains(url, "/") == false
# } else = url {
#   count(data.api[url][method]) > 0
# } else = parent_match {
#     path := split("/",url)
#     numbers.range(1, count(path))


#     num := count(path)
#     last := path[num - 1]

#     parent_api :=  trim_suffix(url, concat("", ["/", last]) )

#     parent_match := longestmatchapi(parent_api, method)
# }


