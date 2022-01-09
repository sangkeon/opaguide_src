package example.basic

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

    some p
    permission[p] == required_permission
}