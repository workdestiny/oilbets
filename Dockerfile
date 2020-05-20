FROM acoshift/go-scratch

ADD poweroil /
COPY template /template
COPY public /public
COPY settings /settings
EXPOSE 8080

ENTRYPOINT ["/poweroil"]