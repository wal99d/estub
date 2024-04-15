
FROM alpine
ADD ./bin/app /app
EXPOSE 2222 
EXPOSE 3000 

ENTRYPOINT [ "/app"]