from grafana/grafana

USER root

RUN apt-get update && apt-get install -y jq
RUN mkdir -p /home/grafana && chown grafana /home/grafana
COPY setup.sh /setup.sh
COPY start.sh /start.sh
COPY dashboard.raw.json /dashboard.raw.json
RUN chmod a+x /start.sh /setup.sh
RUN touch /dashboard.json && chmod a+rwx /dashboard.json

USER grafana

ENTRYPOINT /start.sh
