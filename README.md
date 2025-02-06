# OpenFGA Web Server

openfga_webserver is a Golang-based web server that integrates with OpenFGA to provide fine-grained access control. The project uses Docker Compose to launch both the web server and OpenFGA simultaneously, with the Playground feature enabled for convenient development and testing.

## Project Features

- Fine-grained Access Control
  Leverages OpenFGA to manage and validate user access to resources.

- Routing and Middleware Control
  Custom middleware is used to authenticate requests, ensuring that only authorized users can access protected pages.

- Docker Compose Integration
  Spin up the entire system (web server and OpenFGA) with a single command, simplifying deployment and development environment setup.

- OpenFGA Playground
  The Playground feature is available at http://localhost:3000/playground for viewing and testing the authorization model.

## Pre-configured Users and Access Hierarchy

The project comes pre-configured with three users: Jason, Alice, and Ethan.

- User Hierarchy:

  - Jason can manage Alice.
  - Alice can manage Ethan.

- Personal Pages:
  Each user has their own page. However: - Jason can view everyone's page. - Alice can only view her own page and Ethan's page.

- Testing Access with URLs
  You can experiment with the following URLs to test the access control:

      - Jason viewing Ethan's page
      http://localhost:8000/page/ethan-page?access=jason

      - Ethan viewing Ethan's page
      http://localhost:8000/page/ethan-page?access=ethan

      - Ethan attempting to view Jason's page
      http://localhost:8000/page/jason-page?access=ethan
