
// scripts/app.js
document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form'); // id corretto dal tuo HTML

     loginForm.addEventListener('submit', async function(event) {
        event.preventDefault();
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        // Chiamata PUT al backend
        const response = await fetch('/login', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            const data = await response.json();
            // salva nel sessionStorage
            sessionStorage.setItem('username', data.username);
            sessionStorage.setItem('balance', data.balance);
            sessionStorage.setItem('password', data.password);
            let isFirstLogin = data.firstlogin;
            console.log(isFirstLogin); 

            // Reindirizza alla pagina di set carta
            if (isFirstLogin) {
                window.location.href = '/set-card.html';
            } else {
                const token = sessionStorage.getItem('card_token');
                const expiry = sessionStorage.getItem('card_token_expiry');
                if (!token || !expiry || Date.now() > parseInt(expiry, 10)) {
                    alert('Card token expired or not set. Please set your card again.');
                    window.location.href = '/set-card.html';
                    return;
                }else{

                    window.location.href = '/profile.html';
                }
            }
        } else {
            alert('Login failed!');
        }
    });
 
});