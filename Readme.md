#ftp2s3 + docker

Spawn a ftp server and upload every images onto a S3 bucket

#ftp2s3

The `ftp2s3` binary is a Go program that listens for changes on a given directory, when an image (.png, .jpg or .jpeg) is created in this directory it's uploaded onto a S3 bucket. If the upload is successful, the image is locally destroyed.

###Usage

````bash
export AWS_ACCESS_KEY=
export AWS_SECRET_ACCESS_KEY=
export AWS_BUCKET=applidget-ftp-photo-uploader
ftp2s3 <directory_to_watch> # (default is `.`)
````

It's recommended to work on `ftp2s3` using `vagrant`

#Docker image

The docker image spawns a ftp server (`proftpd`) in the container `/ftp` directory and starts `ftp2s3` on this directory. It allows to get data from devices (camera) which only speak ftp directly from S3. 

###Building the image

1. Build the `ftp2s3` binary (from Vagrant or using GOOS=linux)
2. `cd` into the directory and `docker build .`

###Running the image

Build the image then `docker images` will output images ID. The first one should be the one just built. Then:

````bash
docker run -it -p 21:21 -p 20:20 -p 5000-5100:5000-5100 -e USERNAME=<username> -e PASSWORD=<password> -e AWS_SECRET_ACCESS_KEY=<secret_key> -e AWS_ACCESS_KEY=<access_key> -e AWS_BUCKET=<bucket_name> <image_id>
````

Note port 20 and 21 are basic ftp ports, port 5000 to 5100 are used for ftp passive connection.

To send an image on the ftp server:

````bash
ftp 192.168.59.XX #(or the ip address of boot2docker / your host / your VM )
$> username
$> password
$> put /local/path/to/an/image.png name_on_the_ftp.png
````


 