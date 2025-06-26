curl -X POST http://localhost:8080/api/v1/auth/action/ssh-login \
  -H "Authorization: yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj" \
  -H "Content-Type: application/json" \
  -d '{
    "resource": "aws-cloud-west/server101",
    "login": "support"
  }'
