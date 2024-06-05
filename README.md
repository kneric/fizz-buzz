# fizz-buzz

Run the program using this command
>go run main.go

It will run the server at port `3000` . Use the `GET /range-fizzbuzz` endpoint with these params:
1. `from` -> an integer that is less than or equal to `to`
2. `to`-> an integer 

The endpoint will range the operation from the input above and return:
- Return "Fizz" if `n` is divisible by 3
- Return "Buzz" if `n` is divisible by 5
- Return "FizzBuzz" if `n` is divisible by 3 and 5 
- Return `n` as a default