FROM gcr.io/distroless/static-debian11

ADD bin/kdebug /kdebug

CMD [ "/kdebug" ]
