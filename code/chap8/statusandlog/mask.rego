package system.log

# input의 name 필드가 bob이면 salary 필드 제거
mask["/input/salary"] {
  input.input.name == "bob"
}

# input의 패스워드 필드 무조건 제거
mask["/input/password"]

# input card 필드가 존재하면 갑을 ****-****-****-****로 변경
mask[{"op": "upsert", "path": "/input/card", "value": x}] {
  input.input.card
  x := "****-****-****-****"
}