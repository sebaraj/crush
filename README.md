# Crush

A novel matchmaking service for university students.

As the use of traditional dating apps decline due to social stigma and poor user recommendations,
college students are simultaneously becoming more isolated and less likely to approach potential
romantic interests in person. Crush aims to bridge this gap by providing a platform for students
where they can anonymously express interest in another student, while also receiving partner recommendations based
on their lifestyle, interests, and preferences.

Every week, a user is presented with a curated list of potential matches using a top-n stable
matching algorithm, while also being able to pick one "crush" from a list of all students. If both users express interest in each other, they are notified.

Note: This project is intentionally over-engineered. The Yale College student
body is only 7000 students and the horizontal and vertical scaling of AWS would enable this
to be a simple CRUD app with one RDBMS and S3 bucket. However, I enjoy learning about and
implementing new technologies, and I view this project as a playground for that, while also
providing the foundation for furthing scaling to graduate students and other universities.

### Architecture:

(Archteicture diagram coming soon)

<!-- ![Architecture]() -->

All infrastructure is managed using Terraform as IaC. See `./terraform`

##### Authentication Service

- Stateless service, running on EKS, that uses Google OAuth 2.0 for authentication. The service is also responsible for
  user registration/sign-up (user is auto-populated by Google SSO).

##### User Service

- Stateless service, running on EKS, CRUD operations for users, using PostreSQL for all writes, updates, deletes, and reads on a user's
  own data.
- Opensearch is used for searching for other users based on interests, college, and preferences.
- Generates signed S3 URLs for user profile pictures.

##### S3

- One bucket used for storing website assets, delivered to users via AWS
  Cloudfront/Route53.
- Alternate bucket used to store and deliver profile pictures directly to a user, minimizing load on backend
  services.

##### Match Service

- Stateless service, running on EKS, that handles the reading of matches from the PostgreSQL/RDS
  database, and decoupling the creation and updating of matches through SQS.

##### Match Generation Engine

- Ephermal service, running as a scheduled Lambda, that generates matches based on user preferences and
  interests.
- Uses a top-n variant on the Gale-Shapley algorithm to generate n stable matches per user.
- Utilizes row-level locks to ensure match consistency and correctness on read committed database.

##### SQS + Lambda Consumer

- Decouples user updates on matches and decouples stateless and stateful matching components from the EKS service, allowing for a more scalable, performant, and fault-tolerant
  system.
- Utilizes row-level locks to ensure match consistency and correctness on read committed database.

##### Database

- PostgreSQL 15, hosted on RDS.
- Transaction isolation level: Read Committed
- Schema: see `./database/rds_schema.sql`

##### Opensearch

- Uses change data capture (CDC) via DMS (which itself filters out private user data) to keep the Opensearch index eventually consistent with
  RDS. Avoids unnecessary expenses of 2PC, while still handling 'bursty' search load at scale.

<!-- ### Deploying: -->
<!---->
<!-- - Install dependencies: -->
<!---->
<!--   - Terraform -->
<!---->
<!--   - AWS CLI -->
<!---->
<!--   - kubernetes (kubectl) -->
<!---->

### TO-DO

##### MVP:

- [x] Terraform IaC for AWS

  - [x] VPC
  - [x] EKS
  - [x] RDS
  - [x] S3
  - [x] CloudFront
  - [x] Route53
  - [x] ACM
  - [x] Opensearch, with CDC using DMS
  - [x] Lambda
  - [x] SQS

- [x] Authentication Service, using Google OAuth 2.0
- [x] User Service

  - [x] CRUD with DB
  - [x] Signed S3 URL generation
  - [x] Opensearch integration

- [x] Match Service

  - [x] Match generation engine
  - [x] Read from DB
  - [x] Decoupled Create/Update matches through SQS
  - [x] Lambda SQS consumer

- [ ] Github Actions CI/CD pipelines (Build, Test, Deploy)
  - [x] User, Match, and Authentication services to ECR
  - [ ] SQS Consumer and Match Generation Engine as Lambda

##### Planned Infrastructure:

- [ ] Additional unit tests and more comprehensive integration tests
- [ ] Integrate Logrus for structured logging
- [ ] Security/Permissions:
  - [ ] IAM role/Postgres user with more fine-grained permissions for each service
- [ ] Upgrade to Postgres 17 for additional logical replication features (need DMS
      support; currently not supported)

##### Planned Features:

- [ ] v1.1:
  - [ ] User notification on match via email/text using SNS
  - [ ] Extend to graduate students at Yale
  - [ ] Implement elo based on match rate, considering elo in match generation engine
- [ ] v1.2:
  - [ ] Users can recommend a limited number of matches not involving themselves per week; will be
        factored in by the match generation engine
  - [ ] Users can upload multiple publically-visible photos and a bio to their profile
- [ ] v2.0:
  - [ ] Extend to other universities

##### Other Goals:

- [ ] Terms of Service and Privacy Policy
- [ ] Switch from google forms to native website for user feedback; add a feature updates page

<!-- ### Motivation -->
