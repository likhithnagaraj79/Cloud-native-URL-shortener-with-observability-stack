FROM maven:3.9-amazoncorretto-21 AS builder
WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline -q
COPY src ./src
RUN mvn package -DskipTests -q

FROM amazoncorretto:21-alpine
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=builder /app/target/*.jar /app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app.jar"]
