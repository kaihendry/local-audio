Goal: Proof of concept for a Web Audio walk platform

# Data retention

- dynamdo db "time to live" expires in 1 day from creation of record set in add.go
- s3 lifecycle rule of 1 day in cloudformation template.yml

# Audio capture support in IOS

Apple IOS has broken audio capture device API support: https://twitter.com/anssik/status/1477746208210882562

It works ....but it suprisingly launches the camera and of course uploads
**video** with the audio.

# Start dynamodb server

    ./scripts/local-dynamodb.sh
    ./scripts/create-table.sh

# Start Go Web server

You need to setup the BUCKET_NAME to your own bucket!

    BUCKET_NAME=local-audio-test TABLE_NAME=Records gin --all
