# PayPal-like App

This project is a simple web application that simulates a payment system similar to PayPal. Users can log in, view their balance, and send money to other users. The application is designed for educational purposes, particularly to demonstrate security vulnerabilities such as Cross-Site Scripting (XSS).

## Project Structure

```
paypal-like-app
├── index.html        # Main HTML page for the application
├── styles            # Directory for CSS styles
│   └── style.css     # Styles for the application
├── scripts           # Directory for JavaScript files
│   └── app.js        # JavaScript code for functionality and XSS simulation
└── README.md         # Documentation for the project
```

## Features

- User login with username and password
- Display of user balance
- Functionality to send money to other users
- Demonstration of XSS vulnerabilities

## Setup Instructions

1. Clone the repository to your local machine.
2. Open the `index.html` file in a web browser to view the application.
3. Ensure that you have a local server running if you want to test the JavaScript functionalities.

## Usage Guidelines

- Enter a username and password to log in.
- Once logged in, you can view your balance and send money to other users.
- The application includes a section to simulate an XSS attack for educational purposes. Please use this responsibly and only in a controlled environment.

## Disclaimer

This application is for educational purposes only. Do not use it for real transactions or sensitive data handling. Always follow best practices for web security in production applications.