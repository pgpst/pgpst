# go-spamc

go-spamc is a golang package that connects to spamassassin's spamd daemon. 
Is a code port of nodejs module node-spamc(https://github.com/coxeh/node-spamc) 

Thanks for your amazing code Carl Glaysher ;)



You are able to:

  - Check a message for a spam score and return back what spamassassin matched on
  - Ability to send messages to spamassassin to learn from
  - Ability to do everything that `spamc` is capable of

## Methods Available

  - `Check` checks a message for a spam score and returns an object of information
  - `Symbols` like `check` but also returns what the message matched on
  - `Report` like `symbols` but matches also includes a small description
  - `ReportIfSpam` only returns a result if message is spam
  - `ReportIgnoreWarning` like report but matches only symbols with score > 0 "New"	
  - `Process` like `check` but also returns a processed message with extra headers
  - `Headers` like `check` but also returns the message headers in a array
  - `Learn` abilty to parse a message to spamassassin and learn it as spam or ham
  - `ReportingSpam` ability to tell spamassassin that the message is spam
  - `RevokeSpam` abilty to tell spamassassin that the message is not spam
 

## Example
example.go

    package main
	
    import (
	   "fmt"
	   "spamc"
    )

    func main() {
	
        html := "<html>Hello world. I'm not a Spam, don't kill me SpamAssassin!</html>"
	    client := spamc.New("127.0.0.1:783",10)

	    //the 2nd parameter is optional, you can set who (the unix user) do the call
	    reply, _ := client.Check(html, "saintienn")

	    fmt.Println(reply.Code)
	    fmt.Println(reply.Message)
	    fmt.Println(reply.Vars)
    }



    