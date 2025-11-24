## Prerequisites

Before you begin, ensure you have the following installed on your local machine:

-   [Docker](https://docs.docker.com/get-docker/)
-   [Make](https://www.gnu.org/software/make/) (usually pre-installed on Linux/macOS)

---

## Getting Started

Follow these steps to get the project up and running on your local machine.

### 1. Configure Environment Variables
```sh
# Copy the example .env file to create your own local configuration
cp .env.example .env
```
> Update the  `POST_ENDPOINT` variable with your own endpoint.

### 2. How to Run
```shell
# Build the application's Docker image
make docker-build

# Run the application in a Docker container
make docker-run

# Run lint
make lint
```

### 3. How to use it
> If you changed the .env file, you need to adjust the values in the following command accordingly.
```shell
curl --location 'localhost:8080/log' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 1,
    "total": 1.65,
    "title": "delectus aut autem",
    "meta": {
        "logins": [
            {
                "time": "2020-08-08T01:52:50Z",
                "ip": "0.0.0.0"
            }
        ],
        "phone_numbers": {
            "home": "555-1212",
            "mobile": "123-5555"
        }
    },
    "completed": false
}'
```