Greenlight is a build-along project from the book by Alex Edwards, Let's Go Futher, a sequel to Let's Go. While it is essentially an API service that allows users manage information resource about movies, the goal was to learn and apply core principles to building JSON RESTful APIs in Go with mostly the standard library.

### Run locally

- Clone the repository

    ```bash
    git clone https://github.com/theolujay/greenlight.git
    cd greenlight
    ```

- Create a `.envrc` file with the following:

    ```bash
    export GREENLIGHT_DB_DSN=postgres://<db_user>:<db_password>@localhost/<db_name>
    export SMTP_HOST=<your_smtp_host>
    export SMTP_PORT=<your_smtp_port>
    export SMTP_USERNAME=<your_smtp_username>
    export SMTP_PASSWORD=<your_smtp_password>
    export SMTP_SENDER=<your_smtp_sender>
    ```
- Run it:

    ```bash
    make run/api
    # Run `make help` to see other available commands
    ```
- Explore the API at `localhost:4000`

