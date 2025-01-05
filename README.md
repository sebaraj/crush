# Crush

(description)

<!-- Note: This project is intentionally over-engineered. The Yale College student
body is only 7000 students and the horizontal and vertical scaling of AWS would enable this
to be a simple CRUD app with one RDBMS and S3 bucket. However, I enjoy learning about and
implementing new technologies, and I view this project as a playground for that, while also
providing the Yale student body with a robust and scalable solution. -->

### Architecture:

##### Authentication Service

##### User Service

##### S3

##### Match Service

##### Match Generation Engine

##### SQS + Lambda Consumer

- Generated matches are released bi-monthly at 5 pm, with
  traffic being highly concentrated at that time. To handle this, while ensuring the guarantees
  and restrictions of matching, selecting new matches is decoupled with SQS and searching for
  users is handled via Opensearch (see below).

##### Database

- PostgreSQL 15, hosted on RDS.
- Transaction isolation level: Read Committed
- Schema: see `./database/rds_schema.sql`

##### Opensearch

- Uses change data capture (CDC) via DMS to keep the Opensearch index eventually consistent with
  RDS. Avoids unnecessary expenses of 2PC, while still handling 'bursty' search load at scale.

### Deploying:

- Install dependencies:

  - Terraform

  - AWS CLI

  - kubernetes (kubectl)

### TO-DO

##### MVP:

- [ ] Terraform IaC for AWS

  - [x] VPC
  - [x] EKS
  - [x] RDS
  - [x] S3
  - [x] CloudFront
  - [x] Route53
  - [x] ACM
  - [x] Opensearch, with CDC using DMS
  - [ ] Lambda
  - [ ] SQS

- [x] Authentication Service, using Google OAuth 2.0
- [x] User Service

  - [x] CRUD with DB
  - [x] Signed S3 URL generation
  - [x] Opensearch integration

- [ ] Match Service

  - [ ] Match generation engine
  - [ ] Read from DB
  - [ ] Decoupled Create/Update matches through SQS
  - [ ] Lambda SQS consumer

- [x] Github Actions CI/CD pipelines (Build, Test, Deploy)

##### Planned Infrastructure:

- [ ] Additional unit tests and more comprehensive integration tests
- [ ] Integrate Logrus for structured logging
- [ ] Security/Permissions:
  - [ ] IAM role/PostgresQL user with more fine-grained permissions for each service
- [ ] Upgrade to Postgres 17 for additional logical replication features (need DMS
      support; currently not supported)

##### Planned Features:

- [ ] User notification on match via email/text using SNS
- [ ] Extend to graduate students at Yale

### Motivation
