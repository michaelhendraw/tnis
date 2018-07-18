# readme

Microservices tnis with bank case studies with requirement:
- Customer can make a cash deposit to a bank officer (the bank officer use a mobile app/web).
	.Application user is the bank officer.
	.A cash deposit can be done for new customer to the bank (which don’t have an account in the bank).
	.A cash deposit can also be done for existing customer (which already have an account in the bank).
	.The bank officer can see the history of a customer’s cash deposit transaction to the account.
	.The bank officer can see the total balance of a customer by account number.
- Send a notification / receipt to the customer upon completion of cash deposit. (can be email / sms / push)

Technology :
- Golang Programming Language
- MySQL database
- Elasticsearch

Microservices consist of:

1. tnis-gateway: api gateway to redirect on other microservices

2. tnis-auth: api for authentication and authorization
a) /auth/create: add new client (web, android, ios, another client, etc)
b) /auth/get_token: get token based on client and user
c) /auth/check_token: checking token and expired

3. tnis-customer: api for customer bank data
a) /customer/create: add new customer
b) /customer/update: update customer

4. tnis-transaction: api for transaction saving and withdrawing customer money
a) /transaction/save: save or withdraw money
b) /transaction/history: view customer transaction history
c) /transaction/balance: view latest customer balance

5. tnis-notif: api to send email notification
a) /notif/send_email: send email after transaction saving or withdrawing customer money successfully (auto called after /transaction/save)

Keterangan:
Untuk .sql dan gambaran ERD terlampir pada folder note
Untuk mapping elasticsearch ada pada link postman berikut https://www.getpostman.com/collections/5a0b0beb223f72aa6b53
Untuk mendapatkan app.toml kontak email michaelhendraw@gmail.com
app.toml diletakkan pada setiap folder tnis-.../micin/config

Information:
For .sql and ERD's picture attached to notes folder
For elasticsearch mapping and endpoint is on the following postman link https://www.getpostman.com/collections/5a0b0beb223f72aa6b53
To get app.toml contact email michaelhendraw@gmail.com
app.toml must be placed on every tnis -... /micin/config folder