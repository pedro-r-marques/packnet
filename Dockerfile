FROM ubuntu:14.04

MAINTAINER "Steve Rude <srude@riotgames.com>"

ENV VERSION 2.26

RUN apt-get update -qy
RUN apt-get install -y software-properties-common

# Add OpenContrail repos
RUN add-apt-repository ppa:opencontrail/ppa
RUN add-apt-repository ppa:opencontrail/release-2.01-juno

RUN echo "deb http://ppa.launchpad.net/opencontrail/release-2.01-juno/ubuntu trusty main" >> /etc/apt/sources.list.d/opencontrail.list
RUN echo "deb-src http://ppa.launchpad.net/opencontrail/release-2.01-juno/ubuntu trusty main" >> /etc/apt/sources.list.d/opencontrail.list

RUN echo "deb http://ppa.launchpad.net/opencontrail/ppa/ubuntu trusty main" >> /etc/apt/sources.list.d/opencontrail-deps.list
RUN echo "deb-src http://ppa.launchpad.net/opencontrail/ppa/ubuntu trusty main" >> /etc/apt/sources.list.d/opencontrail-deps.list

RUN apt-get update -qy
RUN apt-get install -y python python-dev python-setuptools python-contrail python-contrail-vrouter-api

# Install nsenter
ADD ./nsenter /usr/bin/nsenter
RUN chmod +x /usr/bin/nsenter

RUN mkdir -p /app
ADD ./scripts /app/scripts
ADD ./packnet /app/packnet
RUN cd /app/scripts/vrouter-ctl && python setup.py install


WORKDIR /app

CMD ["/bin/bash"]