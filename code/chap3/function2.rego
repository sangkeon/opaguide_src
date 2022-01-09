package function2

one("number") = r {
    r := 1
}

one("string") = r {
    r := "one"
}

testRule = result {
    result := one("number")    
}