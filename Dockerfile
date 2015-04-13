FROM ubuntu

RUN apt-get update -qq && apt-get install -y proftpd ca-certificates && mkdir -p /ftp

ADD fswatcher /usr/local/bin/fswatcher

ADD proftpd.conf /etc/proftpd/proftpd.conf
ADD launch.sh /launch.sh

RUN sudo chown root:root /etc/proftpd/proftpd.conf /launch.sh

ENTRYPOINT /launch.sh

