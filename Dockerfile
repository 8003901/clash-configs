#FROM azul/zulu-openjdk:11
FROM openjdk:17
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} clash-configs.jar
RUN sh -c 'touch /clash-configs.jar'
RUN echo "Asia/Shanghai" > /etc/timezone
ENV JAVA_OPTS="-Dserver.port=80"
ENTRYPOINT [ "sh", "-c", "java -XX:MaxRAMPercentage=70.0 $JAVA_OPTS -Djava.security.egd=file:/dev/./urandom -jar /clash-configs.jar" ]
