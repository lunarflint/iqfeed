FROM i386/golang:1.16.0-alpine3.13

WORKDIR /root
COPY go.mod *.go ./
RUN  go build


FROM i386/alpine:latest

ARG IQFEED_INSTALLER_URL='http://www.iqfeed.net/iqfeed_client_6_1_0_20.exe'

RUN \
    wget "$IQFEED_INSTALLER_URL" -O iqfeed_install.exe &&\
    apk add --no-cache wine xvfb-run && winecfg && wineserver --wait &&\
    xvfb-run -s -noreset -a wine iqfeed_install.exe /S && wineserver --wait &&\
    rm iqfeed_install.exe &&\
    wine reg ADD 'HKEY_CURRENT_USER\Software\DTN\IQFeed\Startup' /v ShutdownDelayStartup /t REG_SZ /d 20 /f &&\
    wine reg ADD 'HKEY_CURRENT_USER\Software\DTN\IQFeed\Startup' /v ShutdownDelayLastClient /t REG_SZ /d 0 /f &&\
    wineserver --wait

WORKDIR /root
COPY --from=0 /root/iqfeed ./

EXPOSE 5009 9100 9200 9300 9400

CMD /root/iqfeed