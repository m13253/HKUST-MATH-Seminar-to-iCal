FROM alpine:edge
MAINTAINER StarBrilliant <coder@poorlab.com>

RUN apk add tzdata
ADD x86_64-linux-gnu-HKUST-MATH-Seminar-to-iCal /HKUST-MATH-Seminar-to-iCal
EXPOSE 19777
CMD /HKUST-MATH-Seminar-to-iCal
