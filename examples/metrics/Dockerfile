from debian:stretch

USER root
RUN apt-get update && apt-get install -y curl
COPY start.sh /start.sh
COPY dockmetric /bin/dockmetric
RUN chmod a+x /start.sh

CMD /start.sh
