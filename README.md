# Dockerized IQFeed

# Quick summary
- Alpine based image
- GoLang based application for keeping iqconnect.exe alive
- Optional port tunneling

## Usage
```
docker build -t iqfeed .
docker run\
    -p 5009:5009\
    -p 9100:9100\
    -p 9200:9200\
    -p 9300:9300\
    -p 9400:9400\
    -e IQFEED_PTCLVER=6.1\
    -e IQFEED_PROXY=YES\
    -e IQFEED_LOGINID=XXXXXX\
    -e IQFEED_PASSWD=XXXXXX\
    -e IQFEED_PRODID=XXXXXX\
    -e IQFEED_PRODVER=XXXXXX\
    -it --rm iqfeed
```