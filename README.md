# go-nude

Nudity detection with Go.

(Go porting from https://github.com/pa7/nude.js)

## Example

```go
package main

import (
	"fmt"
	"log"
    "github.com/koyachi/go-nude"
)

func main() {
	imagePath := "images/test2.jpg"

	isNude, err := nude.IsNude(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("isNude = %v\n", isNude)
}

```

## Other implementations

- [nude.js](http://www.patrick-wied.at/static/nudejs/)
- [nude.rb](https://github.com/mitukiii/nude.rb)
- [nude.py](https://github.com/hhatto/nude.py)


## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
