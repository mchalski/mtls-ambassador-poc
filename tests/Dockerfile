FROM python:3.8-slim

RUN pip3 install pytest

RUN mkdir -p /testing
ENTRYPOINT ["bash", "/testing/run.sh"]
