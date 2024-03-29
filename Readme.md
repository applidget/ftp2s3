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

#Docker image development

The docker image spawns a ftp server (`proftpd`) in the container `/ftp` directory and starts `ftp2s3` on this directory. It allows to get data from devices (camera) which only speak ftp directly from S3. 

###Building the image

1. Build the `ftp2s3` binary (from Vagrant or using GOOS=linux)
2. `cd` into the directory and `docker build -t robinmonjo/ftp2s3 .`

###Running the image

Build the image then `docker images` will output images ID. The first one should be the one just built. Then:

````bash
docker run -it -p 21:21 -p 20:20 -p 5000-5100:5000-5100 -e USERNAME=<username> -e PASSWORD=<password> -e AWS_SECRET_ACCESS_KEY=<secret_key> -e AWS_ACCESS_KEY=<access_key> -e AWS_BUCKET=<bucket_name> -e WEB_HOOK=<web_hook> <image_id>
````

###WEB_HOOK

The `WEB_HOOK` env var can be specified. If it exists, `ftp2s3` will post the image URL to the path:

`$WEB_HOOK/buckets/:id/photos.json` where `:id` is replaced by the name of the base folder. For example,
if the ftp server is in `/ftp` and a photo is created in `/ftp/foo/dcim/bar.jpg` we will :

`POST $WEB_HOOK/buckets/foo/photos.json` with body `"remote_photo_url": "http://jqhebkjhbqslqkb"`

Note port 20 and 21 are basic ftp ports, port 5000 to 5100 are used for ftp passive connection.

To send an image on the ftp server:

````bash
ftp 192.168.59.XX #(or the ip address of boot2docker / your host / your VM )
$> username
$> password
$> put /local/path/to/an/image.png name_on_the_ftp.png
````

#In production

## run the image

1. Install docker
2. Make sure host port 21,22 and 5000-5100 are open
3. `sudo docker pull robinmonjo/ftp2s3`
4. Launch the container:

````bash
docker run -p 21:21 -p 20:20 -p 5000-5100:5000-5100 -e USERNAME=<username> -e PASSWORD=<password> -e AWS_SECRET_ACCESS_KEY=<secret_key> -e AWS_ACCESS_KEY=<access_key> -e AWS_BUCKET=<bucket_name> -e WEB_HOOK=<web_hook> -d --restart on-failure:10 --log-driver syslog robinmonjo/ftp2s3
````

5. Logs cat be read with `tail -F /var/log/syslog | grep docker/`

## publish an update

`docker push robinmonjo/ftp2s3`
 