package hello

test_allow_hello_allowed {
    allow_hello with input as {}
}

test_allow_world_not_allowed {
    not allow_world with input as {}
}