# req 

> A http request library for Go.

## Get

```
go get -u -v github.com/hongbook/req
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/hongbook/req"
)

func main() {
	req.SetOptions(
		req.SetBaseURL("https://jsonplaceholder.typicode.com"),
	)

	resp, err := req.Get(context.Background(), "/posts/42", nil)
	if err != nil {
		panic(err)
	}

	body, err := resp.String()
	if err != nil {
		panic(err)
	}
	fmt.Printf("status:%d,body:\n%s\n", resp.Response().StatusCode, body)
}
```

> output:

```
status:200,body:
{
  "userId": 5,
  "id": 42,
  "title": "commodi ullam sint et excepturi error explicabo praesentium voluptas",
  "body": "odio fugit voluptatum ducimus earum autem est incidunt voluptatem\nodit reiciendis aliquam sunt sequi nulla dolorem\nnon facere repellendus voluptates quia\nratione harum vitae ut"
}
```


## MIT License

    Copyright (c) 2020 hongbook

