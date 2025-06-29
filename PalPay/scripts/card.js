// Visualizzazione dinamica della carta
document.addEventListener('DOMContentLoaded', function() {

    function generateToken(length = 32) {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let token = '';
        for (let i = 0; i < length; i++) {
            token += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return token;
    }

    function setCardToken() {
        const token = generateToken();
        const expiresAt = Date.now() + 10 * 60 * 1000; // 10 minuti
        sessionStorage.setItem('card_token', token);
        sessionStorage.setItem('card_token_expiry', expiresAt);
        return;
    }

    document.getElementById('card-number').addEventListener('input', function() {
        let val = this.value.replace(/\D/g, '').replace(/(.{4})/g, '$1 ').trim();
        document.getElementById('visual-card-number').textContent = val.padEnd(16, '•');
    });
    document.getElementById('holder-name').addEventListener('input', function() {
        document.getElementById('visual-holder').textContent = this.value.toUpperCase() || 'NAME SURNAME';
    });
    document.getElementById('expiry').addEventListener('input', function() {
        document.getElementById('visual-expiry').textContent = this.value || 'MM/YY';
    });

    // Submit della carta
    document.getElementById('card-form').addEventListener('submit', async function(e) {
        e.preventDefault();
        const username = sessionStorage.getItem('username');
        if (!username) {
            alert('User not logged in');
            return;
        }
        const card_number = document.getElementById('card-number').value.replace(/\s/g, '');
        const cvv = document.getElementById('cvv').value;
        const expiry = document.getElementById('expiry').value;
        const holder_name = document.getElementById('holder-name').value;

        // Validazione base
        if (!/^\d{16}$/.test(card_number)) {
            alert('Card number must be 16 digits.');
            return;
        }
        if (!/^\d{3}$/.test(cvv)) {
            alert('CVV must be 3 digits.');
            return;
        }
        
        if (!/^\d{2}\/\d{2}$/.test(expiry)) {
            alert('Expiry must be in MM/YY format.');
            return;
        }
        // Validation of the exipre date
        const [month, year] = expiry.split('/').map(Number);
        // Check if month is valid
        if (month < 1 || month > 12) {
            alert('Invalid month in expiry date.');
            return;
        }
        // Check if year is valid
        const currentYear = new Date().getFullYear() % 100; // Get last two digits of the current year
        if (year < currentYear) {
            alert('Card has expired.');
            return;
        }
        
        if (!holder_name.trim()) {
            alert('Holder name required.');
            return;
        }

        const res = await fetch('/set-card', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, card_number, cvv, expiry, holder_name })
        });
        if (res.ok) {
            setCardToken();
            alert('Card saved!');
            window.location.href = '/profile.html';
        } else {
            alert('Error saving card');
        }

    })
});