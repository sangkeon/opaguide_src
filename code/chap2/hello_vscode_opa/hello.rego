package hello

default allow_hello = false

default allow_world = false

allow_hello {
    "hello" != ""
}

allow_world {
    "world" != "world"
}