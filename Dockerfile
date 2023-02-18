FROM gcr.io/distroless/static-debian11

ADD bin/kdebug bin/run-as-host /

CMD [ "/kdebug" ]
