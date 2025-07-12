# The initial version
if [ ! -f .env ]
then
  export $(cat ./.env | xargs)
fi

# My favorite from the comments. Thanks @richarddewit & others!
set -a && source ./.env && set +a

go run cmd/server.go