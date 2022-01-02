Goal: Proof of concept for a Web Audio walk platform

# Data retention

* dynamdo db "time to live" expires in 1 day from creation of record set in add.go
* s3 lifecycle rule of 1 day in cloudformation template.yml

# Start dynamodb server

    ./scripts/local-dynamodb.sh
    ./scripts/create-table.sh

# Start Go Web server

You need to setup the BUCKET_NAME to your own bucket!

    BUCKET_NAME=local-audio-test TABLE_NAME=Records gin --all
