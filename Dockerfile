# Set the base image
FROM debian

# informations
LABEL author="Iceflower S"
LABEL maintainer="Iceflower S"
LABEL email="iceflower@gmx.de"
LABEL version="1.0"
LABEL description="Red Eclipse Development Server"

# we are noninteractive
ENV DEBIAN_FRONTEND noninteractive

# Add server user and set permissions
RUN useradd --create-home --shell /bin/bash redeclipse \
    && mkdir /redeclipse \
    && chown redeclipse: -R /redeclipse

# Update application repository list, create build dir, build server, move server files, create other permanent files and clean up
RUN apt-get -qq update \
    && apt-get install --no-install-recommends -y build-essential zlib1g-dev git ca-certificates cmake pkg-config \
    && git clone -b master https://github.com/red-eclipse/base /temp \
    \
    && mkdir /temp/build \
    && cd /temp/build \
    && cmake ../src -DBUILD_CLIENT=0 \
    && make clean install -j4 \
    \
    && mkdir -p /redeclipse/config \
    && mv /temp/config/ /redeclipse/config/ \
    && mkdir -p /redeclipse/bin/amd64 \
    && mv /temp/bin/amd64/redeclipse_server_linux /redeclipse/bin/amd64/redeclipse_server_linux \
    \
    && apt-get -qq update \
    && apt-get remove --purge -y build-essential zlib1g-dev git ca-certificates cmake pkg-config \
    && apt-get -qq update \
    && apt-get autoremove -y \
    && rm -rf /temp

# Add defaults maps and server config folder
RUN apt-get install --no-install-recommends -y git ca-certificates \
    && git clone -b master https://github.com/red-eclipse/maps.git /redeclipse/data/maps \
    && mkdir -p /home/redeclipse/server-config/ \
    && apt-get remove --purge -y git ca-certificates
    && apt-get -qq update \
    && apt-get autoremove -y

USER redeclipse

# This ports have to be used by the server config
EXPOSE 28804 28805

CMD /redeclipse/bin/amd64 -h/home/redeclipse/server-config/
