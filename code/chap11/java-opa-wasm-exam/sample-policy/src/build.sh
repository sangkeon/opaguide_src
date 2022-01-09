#!/bin/bash
opa build -t wasm -o ../bundle/bundle.tar.gz -e opa/wasm/test/allowed policy.rego data.json
tar xzf ../bundle/bundle.tar.gz  --directory=../wasm /policy.wasm