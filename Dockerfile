FROM debian:jessie

MAINTAINER "Steve Rude <srude@riotgames.com>"

ENV VERSION 2.26

RUN apt-get update -qy
RUN apt-get install -qy python python-setuptools curl build-essential pkg-config

# Install nsenter
RUN mkdir /src
WORKDIR /src
RUN curl https://www.kernel.org/pub/linux/utils/util-linux/v$VERSION/util-linux-$VERSION.tar.gz \
     | tar -zxf-
RUN ln -s util-linux-$VERSION util-linux
WORKDIR /src/util-linux
RUN ./configure --without-ncurses
RUN make LDFLAGS=-all-static nsenter
RUN mv nsenter /usr/bin/nsenter

RUN mkdir -p /app
ADD ./scripts /app/scripts
ADD ./packnet /app/packnet
RUN cd /app/scripts/contrail-vrouter-api && python setup.py install
RUN cd /app/scripts/vrouter-ctl && python setup.py install


WORKDIR /app

CMD ["/bin/bash"]