# trade_engine
Crypto trade engine

Order Matcher is a program that uses an order book to match the buy order (bid) with the sell orders(ask) one or more when the units are matched or match sell order with one or more buy orders to make trades.



#### Approach for designing the order engine:

The engine will listen to the topic and add it to the order book bids queue if it is a buy order else asks queue.

- Processing LIMIT Orders

On every order, the engine will match the existing orders to their counterpart based on the below conditions

- For Buy Order

Process a buy order only when the limit price is less than equal to the current sell price of a security. eg: Current APPLE sell price is 153 and limit price is 150, it will only execute when sell price gets below or equal to 150.
We check the price of every sell order if it is greater than we break else continue to match the order quantities and create trade whenever there is a settlement.
We publish each trade on trades topic on Kafka and the remaining order is pushed back to a queue.

- For Sell Order

We process only when the limit price is greater than the current sell price.
We check the price of every buy order if it is less than the limit price we break else continue to match the order quantities and create trade whenever there is a settlement.
We publish each trade on trades topic and the remaining order is pushed back to a queue.

### Improvements that can be added:
1. A more efficient matching algorithm.
2. Ability to support market order, cancel order, etc as well.
3. Monitoring
4. More Positive and Negative Test cases.

