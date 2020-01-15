# spotifyGroupQueue

[![CircleCI](https://circleci.com/gh/aaronraff/spotifyGroupQueueGo.svg?style=svg)](https://circleci.com/gh/aaronraff/spotifyGroupQueueGo)

The purpose of this project is to simulate a group queue in Spotify. The specific use case that I have built this app for is when you are in a group setting and one person is connected to a speaker and controlling the song queue.

With this web app it is possible to host a room (using your Spotify account) that your friends can then join and queue songs of their choosing.

While this project is an active work in progress, it is in a functional and semi stable state. I will be continuing to improve the code base as well as add new features.

## Todo

- Improve code coverage through tests
- Add more documentation to improve readability

## Development setup

A couple things that you need to do before running the project.

- Register a Spotify application at https://developer.spotify.com

  - This is used to access the Spotify API
  
  - You then need to store your spotify id and secret into environment variables named "SPOTIFY_ID" and "SPOTIFY_SECRET" respectively.
  
  - These values are used when using this Spotify API wrapper: https://github.com/zmb3/spotify
    
- Create a unique session key for the cookie store and store in an environment variable named "SESSION_KEY". This is used in `main.go`

- You will need to run [PostgreSQL](https://www.postgresql.org/) locally and store the connection string in an environment variable named "DATABASE_URL"

- You should also generate a random string and store it in an environment variable named "ENCRYPTION_KEY"
  - This is in `dbutils.go` and is used to encrypt and decrypt Spotify tokens (when storing in the DB)

## Running Locally

First, start PostgreSQL

Then, just run
```
make
```

To run tests, just run
```
make test
```

## Contact

Aaron Raff – [@aaronraff_](https://twitter.com/aaronraff_) on Twitter – aaronraffdev@gmail.com

## Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
