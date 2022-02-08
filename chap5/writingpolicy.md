# Chap 5. Writing Policies Using OPA

## Contents
- Basic API authorization scenarios with OPA
- Improving basic API authorization while adding multiple permissions, public APIs, permission hierarchies, and API hierarchies
- To create and execute test code to test the policies for API authorization 

In this chapter, let's leverage OPA to implement policies that can be applied to real-world applications. First, let's assume aㅜ API authorization policy that can be utilized in real-world applications and write a scenario for that. First of all, basic API access rights are implemented as OPA policies, and then explained by defining various requirements and expanding existing policies.
 
## Basic Scenario
### Requirements
 
The API authorization required for typical systems is assumed at a basic level as follows.

1) End-user, administrator, and operator roles exist in the system.
2) APIs are defined in REST format, and API calling privileges are defined in pairs of {URL, HTTP methods}.
3) If a user logs in, the user ID and role information are recorded in the session.
4) If you have the role of system administrator, you can call the entire API.
5) Operators are granted permission on a permission group basis. A permission group is a group of API call permissions that are related to each other.
6) An operator can call APIs belonging to permission groups granted by an administrator without restriction.
7) When the user calls the API, the API call can be permitted if the target resource is owned by the user.

Taking the actual API call process as an example, the above policy is as follows. Suppose you have access to an API that can query user profiles. For example, suppose /users/{user_id}/profile has a function that looks up the user profile by invoking the GET method. The system administrator can call the entire API, so it can call the API. Operators can only invoke APIs if they have GET privileges at least one of the following URL: /users, /users/{user_id}, /users/{user_id}/profile. Users can invoke APIs that correspond to users/{user_id} because they are associated with the user themselves.
 
For sources related to this scenario, refer to the chap5/basic directory.
 
### Input Schema Definition
Let's define the schema of OPA inputs as the first step to implementing a policy. To determine API privileges, you first need to know about the API that is called now, and you need the user ID and user role information in the user session. It is also necessary to know which users' information is handled by the API being called. To include this information, we define the OPA input schema with the following JSON objects:

```json
{
  "api" : {
    "url" : "<URL of the API called>",
    "method": "<one of 'GET', 'POST', 'PUT', 'DELETE' and 'PATCH'>",
    "target_user_id": "<User ID with API call target in string format, empty string if target is not specific>"
  },
  "user" : {
    "id": "<ID of the user calling the current API>",
    "role": "<Permissions of the user calling the current API, either 'ADMIN', 'OPERATOR' or 'USER'>"
  }
}
```

The OPA does not currently provide functions such as defining and checking input schemas. -> currently input schema supported. Therefore, the OPA input must be checked by an external program and handed over to the OPA or checked within the Lego script. In this book, we assume that validated inputs from external programs are handed over and continue explaining.
 
### Data Definitions
Since the definition of input passed as a factor in each policy decision in OPA has been completed, let's also create data schemas and data that serve as a database for judgment. First, let's define API access.
 
It is defined to be located under the key "api" so that it can be distinguished from other data at the top of API access. Furthermore, under the api key, the URL of each API is keyed to define the access authority object. API URLs can be included in objects and defined as API object arrays, but access to URLs with keys is more efficient than touring in arrays when querying APIs. The OPA official document also recommends defining objects from the perspective of the entire data as objects with url as the key and API information as objects rather than arrays.

```json
{
  "api": {
    "/users/{user_id}/profile" : {
      "GET": ["profile.read"],
      "PUT": ["profile.update"]
    }
  }
}
```

In the example above, the object is defined again with the key "/users/{user_id}/profile" and GET and PUT are defined as the value of the array. This scenario assumes a REST API, so if you need to define permissions other than GET and PUT, you can add POST, DELETE, PATCH, etc. The reason for declaring permissions as an array is to represent situations in which multiple permissions are required to perform a particular API.
 
Next, let's define permissions for the operator. The key to store operator permissions defines "operator_permission" at the same level as the "api" key previously defined. An id is keyed for each operator, and permissions for each operator are stored as an array of strings. An operator with operator1 id has two privileges: profile.read and profile.update, and can access APIs that require those privileges.

```json
"operator_permission" : {
  "oper1" : ["profile.read","profile.update"],
  "oper2" : ["profile.read"]
}
```

The complete file was saved as data.json as follows:

```json
{
  "api": {
    "/users/{user_id}/profile" : {
      "GET": "profile.read",
      "PUT": "profile.update"
    }
  },
  "operator_permission" : {
    "oper1" : ["profile.read","profile.update"],
    "oper2" : ["profile.read"]
  }
}
```

### Writing Policy
Let's write a Rego rule that carries out policy judgments. First of all, the package was named sample.basic because it was the default scenario. A variable indicating whether or not it has access was declared allowed, and false was assigned by default so that it would not be undefined if no rule was satisfied.

```
package example.basic
 
default allowed = false
```

The rules for granting full rights to administrators are as follows: The name of the rule was allowed, the same as the variable, and the input's user object role property was "ADMIN", so that if "ADMIN" then the rule would be true.

```
allowed {
  input.user.role == "ADMIN"
}
```
 
The second rule is a rule that grants permission to the user, and it becomes true if the three conditions as follows are met.

```
allowed {
  input.user.role == "USER"
  input.user.target_user_id != ""
  input.user.target_user_id == input.user.id
}
```

The first condition is that the user object’s  role property of input must be "USER". The second condition checks that the user object target_user_id value in the input is not empty and compares whether the user object id and target_id in the input are the same under the third condition, which is meaningless if they are both empty strings.  The third condition compares the id of the authenticated session with the id of the owner of the resource being processed.

The last rule deals with cases in which the role of the user is an operator.

``` 
allowed {
  input.user.role == "OPERATOR"
  permission := data.operator_permission[input.user.id]
  required_permission := data.api[input.api.url][input.api.method]
 
  some p
  permission[p] == required_permission
}
```

The first line verifies that the role property of the user object in the input is "OPERATOR". The second line uses the user id of the input as the key in the defined data to read the operator's permissions and then assign them to the permission variable. The third line uses the url and method properties of the inputted api object as keys to query the permissions required by api and assign it to the required_permission variable. The line then declares the local variable p. Then, when an element with the same value as required_permission is found while iterating the permission array, the index of that element is assigned to p.

Let's collect all the contents above and save them as policy.rego as follows. The content is simply a continuation of the previous content. Subsequent examples do not introduce the full content, so please download and refer to the source.

```
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
```

### Testing Policy
Let's test whether the policy actually works as wanted. Unit tests for policies can also be written in Rego, but first, let's test permissions for administrators.

Save the following as policy_test.rego.

``` 
package example.basic
 
test_admin_allowed {
  allowed with input as {
    "user": {
      "role" : "ADMIN"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "GET"
    }
  }
}
```

The package was in sample.basic, the same as the policy. First of all, a rule was written to test the administrator's permissions, named test_admin_allowed. OPA recognizes a rule that begins with test_ as a test case. A rule is written with one with statement, which evaluates whether allowed is true when the input is equal to the object after as. Since the criteria for determining administrator recognition were to check whether the role property was "ADMIN", the role attribute was written as ADMIN. The url and method properties of api are specified by randomly selecting one from the data. Other properties, such as id, are irrelevant and are not specified, but it is also a good habit to specify them in a realistic form. Whether the overall test will be sufficient just by specifying the properties can be confirmed by the test coverage described in other parts of the book.
 
Let's test using the opa tool to meet the rules. If the test case files, policy files, and JSON data files are passed to arguments of the opa test command, it shows whether the test case succeeds or fails.

```
$ opa test policy_test.rego policy.rego data.json
PASS: 1/1
```

When running the opa test, the –v option is added to provide more detailed information, including the rule's name and running time.

``` 
$ opa test -v policy_test.rego policy.rego data.json
data.example.basic.test_admin_allowed: PASS (970.2µs)
PASS: 1/1
```

If the role in the test case is changed to AMIN instead of ADMIN, the test will fail.
 
Let's add more test cases to make it a more practical test. This time, let's write a test case to test the operator account. Write the following and add it to the policy_test.rego you created earlier.

```
test_oper1_profile_read_allowed {
  allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper1"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "GET"
    }
  }
}
```

As can be easily expected, the above is a test case that checks whether a user with id is OPER1 and role is OPERATOR is authorized to invoke GET requests for /users/{user_id}/profile APIs. /users/{user_id}/profile is a URL for a user's profile and the method is GET, so you can guess that it is an API that reads the user profile.  The test case was named test_oper1_profile_read_allowed because it is a test that checks whether the user profile can be read by the operator1 user. In the previously written data, the user has rights to the API and should be allowed.

The data written earlier shows that operator1 has PUT privileges as well as GET for /users/{user_id}/profile. Thus, in the case of PUTs, only the method property of api was changed to add a test case named test_oper1_profile_update_allowed.

```
test_oper1_profile_update_allowed {
  allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper1"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "PUT"
    }
  }
}
```

Let's write a test case for oper2 this time. Unlike oper1, which has both read and update privileges to user profiles, oper2 has read privileges only. The test_oper2_profile_read_allowed test cases for read permissions are as follows: It can be seen that other content except user id is the same as the test_oper1_profile_read_allowed test case.

```
test_oper2_profile_read_allowed {
  allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper2"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "GET"
    }
  }
}
```

This time, let's make sure that the oper2 user does not have permission to update the user profile. The test_oper2_profile_update_not_allowed test case used the not statement for the first time. If you want to make sure that the result is false when you write another test case, you can attach a not keyword in front of the door with the test case.

``` 
test_oper2_profile_update_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper2"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "PUT"
    }
  }
}
```

Next time, let's make a test case for the user. The user is granted permission if the user's role attribute is USER and the id and target_id match.

``` 
test_user_allowed {
  allowed with input as {
    "user": {
      "role" : "USER",
      "id" : "user1",
      "target_user_id" : "user1",
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "PUT"
    }
  }
}
```

To create an disallowed test case by changing the target_user_id value of the user object:

``` 
test_user_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "USER",
      "id" : "user1",
      "target_user_id" : "user2",
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "PUT"
    }
  }
}
```

Since we have prepared various test cases for the policy, let's test them to see if they work well. The opa test command, as follows, confirms that all of the tests are successful.

``` 
$ opa test -v policy_test.rego policy.rego data.json
data.example.basic.test_admin_allowed: PASS (971.2µs)
data.example.basic.test_oper1_profile_read_allowed: PASS (0s)
data.example.basic.test_oper1_profile_update_allowed: PASS (0s)
data.example.basic.test_oper2_profile_read_allowed: PASS (0s)
data.example.basic.test_oper2_profile_update_not_allowed: PASS (1.0408ms)
data.example.basic.test_user_allowed: PASS (0s)
data.example.basic.test_user_not_allowed: PASS (0s)
--------------------------------------------------------------------------------
PASS: 7/7
```

To measure more accurate performance, you can add the --bench option when running the test.

```
$ opa test –v –t 10s --bench policy.rego policy_test.rego data.json
data.example.basic.test_admin_allowed 37602 33966 ns/op 24819 timer_rego_query_eval_ns/op
9300 B/op 170 allocs/op
data.example.basic.test_oper1_profile_read_allowed 25687 47053 ns/op 349 timer_rego_external_resolve_ns/op 37582 timer_rego_query_eval_ns/op 12788 B/op 241 allocs/op
data.example.basic.test_oper1_profile_update_allowed 25747 46019 ns/op 117 timer_rego_external_resolve_ns/op 35928 timer_rego_query_eval_ns/op 12788 B/op 241 allocs/op
data.example.basic.test_oper2_profile_read_allowed 25224 46035 ns/op 199 timer_rego_external_resolve_ns/op 36162 timer_rego_query_eval_ns/op 12572 B/op 236 allocs/op
data.example.basic.test_oper2_profile_update_not_allowed 23569 48003 ns/op 170 timer_rego_external_resolve_ns/op 38563 timer_rego_query_eval_ns/op 14405 B/op 258 allocs/op
data.example.basic.test_user_allowed 31176 37074 ns/op 27546 timer_rego_query_eval_ns/op
10259 B/op 186 allocs/op
data.example.basic.test_user_not_allowed 30032 41724 ns/op 32272 timer_rego_query_eval_ns/op 12092 B/op 208 allocs/op
--------------------------------------------------------------------------------
PASS: 7/7
```

The output results are described in order of name of the rule, number of samples, execution time in nanoseconds, time spent evaluating the rule, memory quota in bytes, and memory allocation retrieval. Using benchmark options, the results are informed in nanoseconds by repeated execution of the test, so that the performance can be compared even if the rule is simple and the execution ends quickly.
 
### Tracing queries for debugging
Let's see how the query written is evaluated. If you turn on trace using REPL, you can see how the query is evaluated. Let's read the rules and data into REPL with the opa run command as follows.

```
$ opa run policy.rego data.json
OPA 0.26.0 (commit 62d3900, built at 2021-01-20T18:56:12Z)
 
Run 'help' to see a list of commands and check for updates.

Once the rules and data are loaded, the tracing can be turned on by typing trace as follows: The tracing works as a toggle, so you can turn off the tracing after entering the trace command once more. Assigning an object to input allows you to set inputs for evaluation, followed by querying data.example.basic.allowed to output intermediate trace results.

> trace
> input := {"user":{"role":"ADMIN"}}
Rule 'input' defined in package repl. Type 'show' to see rules.
> data.example.basic.allowed
query:1 Enter data.example.basic.allowed = _
query:1 | Eval data.example.basic.allowed = _
query:1 | Index data.example.basic.allowed (matched 1 rule)
policy.rego:5 | Enter data.example.basic.allowed
policy.rego:6 | | Eval input.user.role = "ADMIN"
policy.rego:5 | | Exit data.example.basic.allowed
query:1 | Exit data.example.basic.allowed = _
query:1 Redo data.example.basic.allowed = _
query:1 | Redo data.example.basic.allowed = _
policy.rego:5 | Redo data.example.basic.allowed
policy.rego:6 | | Redo input.user.role = "ADMIN"
true
```

The tracing results show that when the query was executed, it found a rule that matched the data.example.basic.allowed (which has a value whose body of the rule is true) and satisfied the sixth line role = "ADMIN" in the policy.rego file during the evaluation process. Matching rules are re-evaluated and the result is true.
 
This time, let's find one of the test cases where the evaluation results are false and assign input.

```
> input := {"user":{"role":"OPERATOR","id":"oper2"},"api":{"uri":"/user/{user_id}/profile","method":"PUT"}}
Rule 'input' re-defined in package repl. Type 'show' to see rules.
> data.example.basic.allowed
query:1 Enter data.example.basic.allowed = _
query:1 | Eval data.example.basic.allowed = _
query:1 | Index data.example.basic.allowed matched 0 rules)
policy.rego:3 | Enter data.example.basic.allowed
policy.rego:3 | | Eval true
policy.rego:3 | | Exit data.example.basic.allowed
query:1 | Exit data.example.basic.allowed = _
query:1 Redo data.example.basic.allowed = _
query:1 | Redo data.example.basic.allowed = _
policy.rego:3 | Redo data.example.basic.allowed
policy.rego:3 | | Redo true
false
```

According to the tracing results, no rules matched and entered the line 3 of policy.rego, which is the line that assigned the default value of the rule as false. The final result is false because the line 3 is evaluated finally.

Current tracing of the OPA is lengthy and difficult to grasp at a glance, but can be useful for debugging complex rules.

## Scenarios with multiple API permissions
In the basic scenario, the API required only one permission, and the API could be called if the operator had one of the permissions in the list. Let's define a new scenario with more complexity, and assume that this time the user can have multiple permissions and that the user must have all the necessary permissions to call the API.

For sources related to this scenario, refer to the chap5/multipermission directory.
 
### Data Definitions
Let's create a new directory and create data.json with the following content:

```json
{
  "api": {
     "/users/{user_id}/profile" : {
       "GET": ["profile.read"],
       "PUT": ["profile.read", "profile.update"]
    }
  },
  "operator_permission" : {
    "oper1" : ["profile.read","profile.update"],
    "oper2" : ["profile.read"],
    "oper3" : ["profile.update"]
  }
}
```
 
The definition of a system or input schema is the same as the basic scenario, and the data schema has been changed to reflect the new requirements. The basic definition of the data schema is the same, but the required permission portion has been changed from string to string array.  An operator oper3 with profile.update privileges only is added to test cases where profile.read and profile.update were both required. Looking at the contents of the data, you can see that updating a user profile requires not only updating privileges but also reading privileges.
 
In the basic scenario, if the update permission implicitly included the read permission. But in this scenario, the read permissions and update permissions were completely separated. In some cases, it is necessary to perform only updates through APIs without having to check the existing content.  For example, if the monitoring information is updated periodically to the server, it is not necessary to pre-read the existing values. In this case, it would not be necessary to give the agent permission to read the existing monitoring information. For another example, if the existing user information is divided into details, it may be more desirable to prevent the administrator from reading the user's personal information when updating passwords, etc.
 
### Writing Policy
The package is named example.multipermission, as a scenario allows multiple permissions to be processed for API calls. The rules for checking administrator and user permissions in the basic scenario have not been changed because they do not refer to the data.api. The rules for checking operator privileges are described below. The first step is to check if the role attribute of the user is "OPERATOR", just like the basic scenario.  The permission list was then taken from the operation_permison object of the data using the user id as a key and assigned as a permission variable.  The required_permission variable was then assigned a list of permissions required by the API object using url and method in the api as keys.  In order to check multiple permissions, the elements in the permission array were traversed, and if the two matched against the elements in required permission, the index p of the permission array was added to the set to create a set and assigned to the satisfied_permission_idx variable.  {permission[p] | permission[p] == required_permission[_]} instead of  {p | permission[p] == required_permission[_]} allows you to create a collection of permission strings themselves, not indexes, but does not use values of the permissions themselves, so only smaller indexes are collected and created as a collection. The required_permission's index _ indicates that it is not necessary to store that index value. The next line verifies that the number of permissions requested by api is equal to or greater than zero, because it is meaningless if both are zero, even if the number of permissions required and the number of permissions met is equal. The last line compares the number of required and satisfied permissions, which is a more efficient way to compare whether the permissions actually match each other, because the number of permissions that match the required permissions is counted.

```
package example.multipermisson
 
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
  satisfied_permission_idx := {p | permission[p] == required_permission[_]}
 
  count(required_permission) > 0
  count(satisfied_permission_idx) == count(required_permission)
}
```

### Testing Policy
Same as in the default scenario, permissions for oper1 and oper2 remain unchanged, oper1 has both read and update permissions for user profiles, and oper2 has read permissions only. When updating a user profile, both read and write permissions are required; only read permissions are required to read the profile.

Therefore, it is still the same as the underlying scenario that operator oper1 can read and write both user profiles and oper2 can only read user profiles. So, all the test cases for the basic scenario should be met as it is. Let's copy policy_test.rego to the multipermission scenario directory, then rename the package the same as the newly created policy.rego.
 
As opposed, the newly added operator oper3 must not be able to read the user profile or update the user profile because it only has update permissions on the user profile. Let's add test cases for this.

The contents of the test case confirming that the operator oper3 cannot read the user profile are as follows:

``` 
test_oper3_profile_read_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper3"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "GET"
    }
  }
}
```

The contents of the test case confirming that the operator oper3 cannot update the user profile are as follows:

``` 
test_oper3_profile_update_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper3"
    },
    "api": {
      "url" : "/users/{user_id}/profile",
      "method" : "PUT"
    }
  }
}
```

Let's add the test case to policy_test.rego and then run the test. All newly added tests along with the existing test case can be confirmed to have been successful.

```
$ opa test -v policy.rego policy_test.rego data.json
data.example.multipermisson.test_admin_allowed: PASS (1.0107ms)
data.example.multipermisson.test_oper1_profile_read_allowed: PASS (0s)
data.example.multipermisson.test_oper1_profile_update_allowed: PASS (0s)
data.example.multipermisson.test_oper2_profile_read_allowed: PASS (999.7µs)
data.example.multipermisson.test_oper2_profile_update_not_allowed: PASS (0s)
data.example.multipermisson.test_oper3_profile_read_allowed: PASS (0s)
data.example.multipermisson.test_oper3_profile_update_not_allowed: PASS (0s)
data.example.multipermisson.test_user_allowed: PASS (998.7µs)
data.example.multipermisson.test_user_not_allowed: PASS (0s)
--------------------------------------------------------------------------------
PASS: 9/9
```

## Scenario with Public APIs
For the service, not only the API that is controlled by authorization but also the public API that can be called without authorization is required. Let's add a public API to the multi-permission scenario.
 
For sources related to this scenario, refer to the chap5/publicapi.

### Data Definitions
Assigning a list of required permissions to an empty array allows public APIs to be represented, so there is no need to change the data structure. The required permissions for testing added GET requests for /about URLs as follows: Let's create a directory for the public API scenario and save it as data.json.
 
{
  "api": {
    "/users/{user_id}/profile" : {
       "GET": ["profile.read"],
       "PUT": ["profile.read", "profile.update"]
    },
    "/about" : {
      "GET" : []
    }
  },
  "operator_permission" : {
    "oper1" : ["profile.read","profile.update"],
    "oper2" : ["profile.read"],
    "oper3" : ["profile.update"]
  }
}

### Writing Policy
For public API scenarios, copy the policy.rego of the multi-permission scenario to the newly created directory and rename the package to example.publicapi. Public APIs are APIs with zero permissions required when importing api from data into url and method properties. If we move the rule as it is, let's add it to policy.rego.

``` 
allowed {
  required_permission := data.api[input.api.url][input.api.method]
 
  count(required_permission) == 0
}
```

### Testing Policy
Let's check the test case to see if we can access the newly added API /about through the public API even if we don't have permission. Let's create a test case that fails if you check permissions as follows and change the url and method properties of api to "/about" and "GET" respectively. In this test case, users must have the same target_user_id attributes to have API permissions, but they do not have any API calling permissions because they do not have the target_user_id attribute. Therefore, it is a test case that tests whether /about can be called to GET without permission.

``` 
test_user_public {
  allowed with input as {
    "user": {
      "role" : "USER",
      "id" : "user1",
    },
    "api": {
      "url" : "/about",
      "method" : "GET"
    }
  }
}
```

Let's add that test case to the policy_test.rego copied from the existing multi-permission scenario, rename the package to example.publicapi, and run the test. Running the test confirms that both the existing test case and the newly added test case were successful as follows:

``` 
$ scenario\publicapi>opa test -v policy.rego policy_test.rego data.json
data.example.publicapi.test_admin_allowed: PASS (999.7µs)
data.example.publicapi.test_oper1_profile_read_allowed: PASS (0s)
data.example.publicapi.test_oper1_profile_update_allowed: PASS (1.002ms)
data.example.publicapi.test_oper2_profile_read_allowed: PASS (0s)
data.example.publicapi.test_oper2_profile_update_not_allowed: PASS (0s)
data.example.publicapi.test_oper3_profile_read_not_allowed: PASS (978.7µs)
data.example.publicapi.test_oper3_profile_update_not_allowed: PASS (0s)
data.example.publicapi.test_user_allowed: PASS (0s)
data.example.publicapi.test_user_not_allowed: PASS (989µs)
data.example.publicapi.test_user_public: PASS (0s)
```

## Scenarios with permission hierarchy
This time, let's add hierarchies of permission to the scenario. For example, if you define all permissions to a user profile as a single permission, a profile, you can treat it as having both profile.read and profile.update to update the user profile. When defining permissions, it would be easier to define complex permissions if these hierarchies could be defined.
 
For sources related to this scenario, refer to the chap5/permissionhierarchy.

### Data Definitions
In the data used in the public API scenario, oper 1 had the two permissions, profile.read and profile.update and In this scenario oper1 now have a single permission - profile. If permission hierarchy is supported, existing test cases must be successful without changing the test case.

```
{
  "api": {
    "/users/{user_id}/profile" : {
      "GET": ["profile.read"],
      "PUT": ["profile.read", "profile.update"]
    },
    "/about" : {
      "GET" : []
    }
  },
  "operator_permission" : {
    "oper1" : ["profile"],
    "oper2" : ["profile.read"]
  }
}
```

### Writing Policy
To support the permission  hierarchy, only operator-related rules in existing scenarios were modified as follows: In previous scenarios, we simply compared the required and current permissions to the same.  The new permissionmatch rule additionally checks if the required permission starts with a string which concatenates an owned permission and the hierarchy separator(in this case dot). The reason why it does not simply compare whether it starts with ownership is because, for example, the privilege "profiletest" and the privilege "profile" do not contain each other, but meet the conditions. Therefore, if the user has the right to profile and the right to inspect is profile.read, the profile must be compared with a . separating the hierarchy to obtain accurate results. These logic works well not only in profile.read but also in multiple hierarchies such as profile.read.name.

``` 
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
```
 
### Testing Policy
Since the test case has not changed, we copied it from policy_test.rego in the public API directory and changed only the package to example.permissionhierarchy. Testing confirms that all previous test cases were successful.

```
$ opa test -v policy.rego policy_test.rego data.json
data.example.permissionhierarchy.test_admin_allowed: PASS (971.2µs)
data.example.permissionhierarchy.test_oper1_profile_read_allowed: PASS (0s)
data.example.permissionhierarchy.test_oper1_profile_update_allowed: PASS (999.7µs)
data.example.permissionhierarchy.test_oper2_profile_read_allowed: PASS (0s)
data.example.permissionhierarchy.test_oper2_profile_update_not_allowed: PASS (0s)
data.example.permissionhierarchy.test_oper3_profile_read_not_allowed: PASS (1ms)
data.example.permissionhierarchy.test_oper3_profile_update_not_allowed: PASS (0s)
data.example.permissionhierarchy.test_user_allowed: PASS (0s)
data.example.permissionhierarchy.test_user_not_allowed: PASS (0s)
data.example.permissionhierarchy.test_user_public: PASS (999.9µs)
--------------------------------------------------------------------------------
PASS: 10/10
```

## Scenario with API hierarchy
Supporting the API URL hierarchy as well as the permission hierarchy will make it easier to manage complex APIs. Let's define the API hierarchy to be useful for more practical applications, although it may look similar to the permission hierarchy.

Basically,If the API URL is in the same form as /a/b/c, let's define permission on /a implicitly includes permission of /a/b or /a/b/c, and etc. This is equivalent to the hierarchical structure of the permission, except for the difference separators( . for permission and / for API). However, in some cases, a sub-API must be implicitly accessible with super-API permission  and some APIs have independent permissions regardless of the URL hierarchy.
To make these two applicable at the same time, /a/b, /a/b/c. a/b/c/d, etc. can all override required permissions with permissions of /a/b, if the permissions of /a are defined separately.

That is, if there are no directly defined permissions , follow the parent permissions; if there is a direct definition, follow the defined permission; If multiple API permission hierarchies are defined at the parent, follow the most specific parent permission . In other words, it follows the hierarchy of permission, but can be overridden at a more specific level.

For sources related to this scenario, refer to the  chap5/urlhierarchy.

### Data definitions
To test the API hierarchy, we defined permissions for the /users URL, the parent URL of the existing /users/{user_id}/profile. The reason why permissions on /users are defined is to verify that permissions of /users are checked on /users requests, /users/{user_id}/profile requests are checked on /users/{user_id}/profile requests, and permissions of /users are checked on undefined /users/{user_id} reqeuests. 

```json
{
  "api": {
    "/users/{user_id}/profile" : {
      "GET": ["profile.read"],
      "PUT": ["profile.read", "profile.update"]
    },
    "/users" : {
      "GET": ["user.read"],
      "PUT": ["user.read", "user.update"]
    },
    "/about" : {
      "GET" : []
    }
  },
  "operator_permission" : {
    "oper1" : ["profile"],
    "oper2" : ["profile.read"],
    "oper3" : ["profile.update"],
    "useroper": ["user.read", "profile.read"]
  }
}
```

### Wring Policy
The policy codes that support API hierarchy are as follows: The package is named example.urlhierarchy, and administrators and users do not use API information to determine permissions, so it is the same as before.  The most important difference is that the url used to assign required_permission is not simply the url property of the api object, but the evaluation result of the longestmatchapi rule. The logestmatchapi rule finds the most specific url applicable to that url, as the name suggests. For example, to check the permissions of "/users/{user_id}", if the permissions for "/users/{user_id}" are defined, return "/users/{user_id}", if not, find if "/users" defined. In this way, we explore the url hierarchy and return the most specific url that matches. API permission checks the url and method properties together, so when checking the parent url, they actually check if the method matches.

```
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
```

The contents of the longestmatchapi rule are described below. First of all, the first line divides the string into / delimiters. Next, an array is created from the length of the previously divided array to 1 and assigned as ns. For example, if the array length is 3, it will be [3,2,1].  The last line is a bit complicated, but it is easy to understand from the concat ("/", array.slice (urlslice, 0, ns[_])) section first. If ns is [3,2,1] due to ns[_], it is repeated for concat ("/", array.slice (urlslice, 0, 3), concat ("/", array.slice (urlslice, 0, 2), concat ("/", array.slice (urlslslice, 0, 1). If the url argument of the rule is "/users/{user_id}/profile", the urlslice becomes ["", "users", "{user_id}", "profile"]. Then array.slice(0,3) becomes ["", "users", "{user_id}", and "profile". In addition, array.slice(0, 2) becomes ["", "users", "{user_id}"]. array.slice (urlslice, 0, 1) becomes ["", "users", "]. When connected back to concat, it becomes "/users/{user_id}/profile", "/users/{user_id}", and "/users", respectively. 

Then, assign a concatenated string to an api variable, use the api and method argument to examine whether an element exists in an api object, and, if present, collect only the api variables present in an array comprehension,  then assign the first element to an apimatch.
= apimatch part in the declaration of rules makes the apimatch variable assigned to the result of the rule.

A closer look at the rules shows that longestmatchapi is also applied to rules declaring public APIs. A line with only url in that rule serves to check whether the result of longestmatchapi exists, not undefined. Whether to apply the API hierarchy to public APIs is a matter of choice, but in this example, we chose to apply it.

If you look at the logic of the longestmatchpi rule, you might think it is a little inefficient. You might think that it would be more efficient to return only one matching in the first place without storing all matching api urls in an array. In the process of writing this example, we have tried, and the for loop and break and continue statements of the traditional programming language are not supported in Rego, making it difficult to implement. I also tried this because I thought it would be possible to implement it recursively using a function, but it failed because recursions of rules are not supported in the Rego(the function is also a rule in Rego).
Introducing logic written in recursive form is as follows. In the case of some language, if the function is called as longestmatchapi -> longestmatchapi2 -> longestmatchapi -> … form, not recognized as recursive and allowed to be invoked in this way, depending on the language, but in the case of Rego, it was also detected as recursive. It is a code that cannot be executed, so we will not explain it in more detail, and if you are interested, please take a look.

```
longestmatchapi(url, method) = "" {
  contains(url, "/") == false
} else = url {
  count(data.api[url][method]) > 0
} else = parent_match {
  path := split("/",url)
  num := count(path)
  last := path[num - 1]
 
  parent_api := trim_suffix(url, concat("", ["/", last]) )
 
  parent_match := longestmatchapi(parent_api, method)
}
```

In the longestmatchpi rule, all of the url pieces that can be matched are arranged and may be considered extremely inefficient. The size of the array is equal to the size of the array divided by the / by the separator. If the URL hierarchy of the API is not too deep, it is not a serious problem for performance.  For more efficient implementation, implementing it as an OPA built-in function may be necessary, but implementing an OPA built-in function requires management of built-in functions, so we recommend implementing it only if performance is a problem.

### Testing Policy
Since we have finished writing the policy, let's fill out the test case. Let's copy policy_test.rego from the previous scenario and rename the package to example.urlhierarchy. First, let's add the following test cases to test permissions for the newly added /users API and useroper operators.

``` 
test_useroper_userlist_read_allowed {
  allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "useroper"
    },
    "api": {
      "url" : "/users",
      "method" : "GET"
    }
  }
}
```

When checking permissions on /users/{user_id}, the permissions of /users must be applied by exploring the API hierarchy because they are not defined in the data. Therefore, the following test cases were added to check whether permissions were allowed on /users/{user_id} as permissions were granted on /users, GET by the operator useroper:

``` 
test_useroper_user_read_allowed {
  allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "useroper"
    },
    "api": {
      "url" : "/users/{user_id}",
      "method" : "GET"
    }
  }
}
```

The operator useroper should not be allowed to attempt a PUT because it can only GET for /users and cannot do a PUT. The test cases for this are as follows.

```
test_useroper_user_update_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "useroper"
    },
    "api": {
      "url" : "/users/{user_id}",
      "method" : "PUT"
    }
  }
}
```

The operator useroper has GET access to /users, but the operator oper1 does not have permission to /users. Therefore, the operator oper1 does not have permission for /users/{user_id}. This is expressed as a test case as follows.

```
test_oper1_user_read_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper1"
    },
    "api": {
      "url" : "/users/{user_id}",
      "method" : "GET"
    }
  }
}
```

Access to APIs with no API hierarchy specified in the data should not be allowed. For example, /goods has no hierarchy with APIs defined in the data. If you express this as a test case, it is as follows.

``` 
test_goods_not_allowed {
  not allowed with input as {
    "user": {
      "role" : "OPERATOR",
      "id" : "oper2"
    },
    "api": {
      "url" : "/goods",
      "method" : "PUT"
    }
  }
}
```

Let's run the whole test case. Running the test confirms that all 10 existing test cases and newly added test cases are successful.

``` 
$ opa test -v policy.rego policy_test.rego data.json
data.example.urlhierarchy.test_admin_allowed: PASS (2.9694ms)
data.example.urlhierarchy.test_oper1_profile_read_allowed: PASS (1.0002ms)
data.example.urlhierarchy.test_oper1_profile_update_allowed: PASS (2.9995ms)
data.example.urlhierarchy.test_oper2_profile_read_allowed: PASS (1.0251ms)
data.example.urlhierarchy.test_oper2_profile_update_not_allowed: PASS (1.0055ms)
data.example.urlhierarchy.test_oper3_profile_read_not_allowed: PASS (1.9995ms)
data.example.urlhierarchy.test_oper3_profile_update_not_allowed: PASS (999.8µs)
data.example.urlhierarchy.test_user_allowed: PASS (0s)
data.example.urlhierarchy.test_user_not_allowed: PASS (971.5µs)
data.example.urlhierarchy.test_user_public: PASS (0s)
data.example.urlhierarchy.test_useroper_userlist_read_allowed: PASS (2.9992ms)
data.example.urlhierarchy.test_useroper_user_read_allowed: PASS (1.0002ms)
data.example.urlhierarchy.test_useroper_user_update_not_allowed: PASS (999.8µs)
data.example.urlhierarchy.test_oper1_user_read_not_allowed: PASS (1.0001ms)
data.example.urlhierarchy.test_goods_not_allowed: PASS (0s)
--------------------------------------------------------------------------------
PASS: 15/15
```

## Summary
In this chapter an API authorization scenario similar to that used in real-world development is assumed, and implements this scenario as an OPA policy. Policies were developed from the most basic scenario and multiple permission, public APIs, permission hierarchies, and API hierarchies are added. In addition, the process of progressively developing the policy and adding test cases while maintaining the tests developed in the previous phase was described step by step.

While the scenarios described in this chapter may be far from the reader's actual requirements, a good understanding of these examples will greatly help to create policies through OPA, as they actually demonstrate the level of complexity and patterns required in developing applications and services.
