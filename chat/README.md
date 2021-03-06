# Twilio Programmable Chat

## Requirements

- [Go](https://golang.org/doc/install) 1.13

## Usage

```go
import (
    "github.com/twilio-go/"
    twchat "github.com/twilio-go/chat"
)
func main() {
    // Configuration required for twilio services
    configuration := twilio.NewContext(accSID, authToken, region)

    // Chat client
    chat, err := twchat.New(configuration)
    if err != nil {
        log.Fatal(err)
    }

    // Create a new service
    // Post https://chat.twilio.com/v2/Services
    service, err := chat.Services.Create(ctx, ServiceCreateParams{"chat-v1"})
    if err != nil {
        log.Fatal(err)
    }

    // Read a service
    // Get https://chat.twilio.com/v2/Services/ISXXXXXX
    service, err := chat.Services.Read(ctx, service.SID)
    if err != nil {
        log.Fatal(err)
    }

    // Update an existing service
    // Post https://chat.twilio.com/v2/Services/ISXXXXX
    service, err = chat.Services.Update(ctx, service.SID)
    if err != nil {
        log.Fatal(err)
    }

    // Delete an existing service
    // Delete https://chat.twilio.com/v2/Services/ISXXXXX
    err = chat.Services.Delete(ctx, service.SID)
    if err != nil {
        log.Fatal(err)
    }
}
```