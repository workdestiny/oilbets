FROM acoshift/go-scratch

ADD amlp /
COPY template /template
COPY public /public
COPY settings /settings
EXPOSE 8080

ENTRYPOINT ["/amlp"]