const { loadPolicy } = require("@open-policy-agent/opa-wasm");
const fs = require('fs');

const policyWasm = fs.readFileSync('policy.wasm');
const dataJson = fs.readFileSync('data.json','utf-8');

input = JSON.parse('{"user": "alice"}');
data = JSON.parse(dataJson);

loadPolicy(policyWasm).then(policy => {
    policy.setData(data);
    resultSet = policy.evaluate(input);
    if (resultSet == null) {
        console.error("evaluation error")
    }
    if (resultSet.length == 0) {
        console.log("undefined")
    }
    console.log("resultSet=" + JSON.stringify(resultSet))
    console.log("result=" + resultSet[0].result) 
}).catch( error => {
    console.error("Failed to load policy: ", error);
})