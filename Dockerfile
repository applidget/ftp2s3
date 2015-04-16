FROM ubuntu

RUN apt-get update -qq && apt-get install -y proftpd ca-certificates && mkdir -p /ftp/before0 /ftp/before1 /ftp/before2 /ftp/before3 /ftp/before4 /ftp/before5 /ftp/before6 /ftp/before7 /ftp/before8 /ftp/before9

ADD ftp2s3 /usr/local/bin/ftp2s3

ADD proftpd.conf /etc/proftpd/proftpd.conf
ADD launch.sh /launch.sh

RUN sudo chown root:root /etc/proftpd/proftpd.conf /launch.sh

ENTRYPOINT /launch.sh

