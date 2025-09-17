# Quickstart: Browser-Based Interactive CLI Server

## Backend Setup
1. Navigate to the project root.
2. Run `go mod tidy` to install dependencies.

## Frontend Setup
1. Navigate to the `frontend` directory.
2. Run `npm install` to install dependencies.

## Running the Server
1. From the project root, run `go run ./cmd/judo server`.
2. This will start the backend server and automatically open the frontend in your browser.

## Building for Production
1. From the project root, run `./build.sh` to build both frontend and backend with embedded assets.
2. This will create a `judo` binary with the frontend assets embedded and run tests.

### Manual Build (Alternative)
1. From the `frontend` directory, run `npm run build`.
2. This will create a `build` directory with the compiled frontend assets.
3. From the project root, run `go build -o judo ./cmd/judo`.
4. This will create a `judo` binary with the frontend assets embedded.
