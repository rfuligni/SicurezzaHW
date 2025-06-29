
document.addEventListener('DOMContentLoaded', function() {
;   
    
    


    function checkCardToken() {
        const token = sessionStorage.getItem('card_token');
        const expiry = sessionStorage.getItem('card_token_expiry');
        if (!token || !expiry || Date.now() > parseInt(expiry, 10)) {
            return true
        }else {
            return false;
        }
    }

    

    if (checkCardToken()) {
        alert('Session expired: Please log in and set your card again for security reasons.');
        window.location.href = '/login.html';
        return;
    }

    async function updateBalanceFromBackend() {
        const username = sessionStorage.getItem('username');
        if (!username) return;
        const res = await fetch(`/get-balance?username=${encodeURIComponent(username)}`);
        if (res.ok) {
            const data = await res.json();
            sessionStorage.setItem('balance', data.balance);
            document.getElementById('balance').textContent = data.balance;
        }
    }

    document.getElementById('username').textContent = sessionStorage.getItem('username') || 'Guest';
    updateBalanceFromBackend();
    document.getElementById('balance').textContent = sessionStorage.getItem('balance') || '0';

    // Load transactions on page load
    loadTransactions();
    document.getElementById('logout').addEventListener('click', function() {
        sessionStorage.clear();
        window.location.href = '/login.html'; // Reindirizza alla pagina di login
    });

    // Modal logic
    const modalOverlay = document.getElementById('modal-overlay');
    const addFundsBtn = document.getElementById('add-funds');
    const closeModalBtn = document.getElementById('close-modal');
    const addFundsForm = document.getElementById('add-funds-form');

    addFundsBtn.addEventListener('click', async function() {
        modalOverlay.classList.add('active');
            // Carica la carta associata
        const username = sessionStorage.getItem('username');
        const cardContainer = document.getElementById('user-card-container');
        cardContainer.innerHTML = '<p>Loading card...</p>';
        try {
            const res = await fetch(`/get-card?username=${encodeURIComponent(username)}`);
            if (!res.ok) {
                cardContainer.innerHTML = '<p style="color:red;">No card found. Please set your card first.</p>';
                return;
            }
            const card = await res.json();
            // Oscura il numero carta (es: 1234 **** **** 5678)
           const showCard = card.card_number.replace(/^(\d{4})\d{8,10}(\d{4})$/, '$1 **** **** $2');
            cardContainer.innerHTML = `
                <div style="display: flex; align-items: center; justify-content: center; width: 100%; margin-bottom: 16px;">
                    <span id="select-dot" style="display:inline-block; width:10px; height:10px; border-radius:500%; border:2.5px solid #007bff; margin-right:2px; background:#fff; transition: background 0.2s, border 0.2s;"></span>
                    <div id="selectable-card" style="border:2px solid #ccc; border-radius:14px; background:linear-gradient(135deg,#4e54c8,#8f94fb); color:#fff; width:260px; padding:18px 16px; cursor:pointer; box-shadow:0 2px 8px rgba(0,0,0,0.12); transition: border 0.2s;">
                        <div style="font-size:1.1em; letter-spacing:2px; margin-bottom:10px; white-space:nowrap;">${showCard}</div>
                        <div style="display:flex; justify-content:space-between; font-size:0.95em;">
                            <div>
                                <div style="font-size:0.8em;">Card Holder</div>
                                <div style="font-weight:bold;">${card.holder_name.toUpperCase()}</div>
                            </div>
                            <div>
                                <div style="font-size:0.8em;">Expires</div>
                                <div>${card.expiry}</div>
                            </div>
                        </div>
                    </div>
                </div>
                <button id="confirm-add-funds" style="background:#28a745; color:#fff; border:none; border-radius:18px; padding:8px 0; width:80%; font-size:15px; margin:12px auto 0 auto; display:block; cursor:pointer;" disabled>Use this card</button>
                <div id="amount-section" style="display:none; margin-top:12px;">
                    <label for="amount-input" style="color:#333;">Amount</label>
                    <input type="number" id="amount-input" min="1" required placeholder="Amount" style="padding:7px; border-radius:4px; border:1px solid #ccc; width:100%; margin-bottom:8px;">
                    <button id="final-confirm" style="background:#007bff; color:#fff; border:none; border-radius:10px; padding:8px 0; width:80%; font-size:15px; margin:0 auto; display:block; cursor:pointer;">Confirm Add Funds</button>
                </div>
            `;

            const selectableCard = document.getElementById('selectable-card');
            const selectDot = document.getElementById('select-dot');
            const confirmBtn = document.getElementById('confirm-add-funds');
            const amountSection = document.getElementById('amount-section');

            let cardSelected = false;

            selectableCard.addEventListener('click', function() {
                cardSelected = !cardSelected;
                if (cardSelected) {
                    selectableCard.style.border = '2.5px solid #28a745';
                    selectDot.style.background = '#28a745';
                    selectDot.style.border = '2.5px solid #28a745';
                    confirmBtn.disabled = false;
                    confirmBtn.focus();
                } else {
                    selectableCard.style.border = '2px solid #ccc';
                    selectDot.style.background = '#fff';
                    selectDot.style.border = '2.5px solid #007bff';
                    confirmBtn.disabled = true;
                }
            });

            confirmBtn.addEventListener('click', function() {
                if (!cardSelected) return;
                amountSection.style.display = 'block';
            });

            document.getElementById('final-confirm').addEventListener('click', async function() {
                const amount = parseFloat(document.getElementById('amount-input').value);
                if (isNaN(amount) || amount <= 0) {
                    alert('Amount must be a positive number.');
                    return;
                }
                // Aggiorna balance nel backend
                const response = await fetch('/add-funds', {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, amount: parseFloat(sessionStorage.getItem('balance')) + amount })
                });
                if (response.ok) {
                    const data = await response.json();
                    sessionStorage.setItem('balance', data.balance);
                    document.getElementById('balance').textContent = data.balance;
                    modalOverlay.classList.remove('active');
                    
                    alert('Funds added!');
                    window.location.reload(); // Ricarica la pagina per aggiornare le transazioni
                } else {
                    alert('Error adding funds.');
                }
            });
            } catch (e) {
                cardContainer.innerHTML = '<p style="color:red;">Error loading card.</p>';
            }
    });

    closeModalBtn.addEventListener('click', function() {
        modalOverlay.classList.remove('active');
    });

    // Optional: close modal clicking outside
    modalOverlay.addEventListener('click', function(e) {
        if (e.target === modalOverlay) {
            modalOverlay.classList.remove('active');
        }
    });

    
    


    // Send money logic
    const sendMoneyModalOverlay = document.getElementById('send-money-modal-overlay');
    const sendMoneyBtn = document.getElementById('send-money');
    const closeSendMoneyModalBtn = document.getElementById('close-send-money-modal');
    const sendMoneyForm = document.getElementById('send-money-form');

    sendMoneyBtn.addEventListener('click', function(){
        sendMoneyModalOverlay.classList.add('active');
    });

    closeSendMoneyModalBtn.addEventListener('click', function() {
        sendMoneyModalOverlay.classList.remove('active');
    });

    sendMoneyModalOverlay.addEventListener('click', function(e) {
        if (e.target === sendMoneyModalOverlay) {
            sendMoneyModalOverlay.classList.remove('active');
        }
    });

    sendMoneyForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const receiver = document.getElementById('receiver-username').value;
        const amount = parseFloat(document.getElementById('send-amount').value);
        const message = document.getElementById('send-message').value;
        const sender = sessionStorage.getItem('username');
        const currentBalance = parseFloat(sessionStorage.getItem('balance'));

        // Validate input
        if (amount > currentBalance) {
            alert('Amount exceeds your current balance.');
            return;
        }
        if (amount <= 0 || isNaN(amount)) {
            alert('Please enter a valid amount to send.');
            return;
        }
        // Send transaction to backend
        const response = await fetch('/send-money', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ sender, receiver, amount, message })
        });

        if (response.ok) {
            const data = await response.json();
            sessionStorage.setItem('balance', data.newBalance);
            document.getElementById('balance').textContent = data.newBalance;
            updateBalanceFromBackend(); // Update balance in the UI
            // Close the modal
            sendMoneyModalOverlay.classList.remove('active');

            alert('Money sent successfully!');
            window.location.reload(); // Reload the page to update transactions
        } else {
            alert('Error sending money.');
        }

    });




    async function loadTransactions() {
        const username = sessionStorage.getItem('username');
        
        if (!username) {
            alert('User not logged in.');
            return;
        }

        try {
            const response = await fetch(`/transactions?username=${username}`);
            if (!response.ok) {
                throw new Error('Failed to fetch transactions.');
            }

            const transactions = await response.json();
            const movementsList = document.getElementById('movements-list');
            movementsList.innerHTML = ''; // Clear existing transactions

            if (transactions === null) {
                // If no transactions, display a message
                const noTransactionsMessage = document.createElement('li');
                noTransactionsMessage.textContent = 'No transactions yet';
                noTransactionsMessage.style.fontStyle = 'italic';
                movementsList.appendChild(noTransactionsMessage);
                return;
            }

            transactions.forEach(transaction => {
                const transactionItem = document.createElement('li');
                transactionItem.style.background = "#fff";
                transactionItem.style.margin = "10px";
                transactionItem.style.marginRight = "25px";
                transactionItem.style.marginLeft = "0px";
                transactionItem.style.padding = "14px 16px";
                transactionItem.style.borderRadius = "12px";
                transactionItem.style.boxShadow = "0 2px 8px rgba(78,84,200,0.08)";
                transactionItem.style.display = "flex";
                transactionItem.style.flexDirection = "column";
                transactionItem.style.fontFamily = "'Segoe UI', Arial, sans-serif";
                transactionItem.style.fontSize = "1.05em";
                transactionItem.style.color = "#333";

                // format timestamp
                const date = new Date(transaction.timestamp);
                const formattedTimestamp = `${date.getDate()}/${date.getMonth() + 1}/${date.getFullYear()} ${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;

                let mainText = "";
                let messageText = "";

                if (transaction.sender === transaction.receiver) {
                    mainText = `<span style="color:#4e54c8;font-weight:600;">Deposit</span> <span style="color:#28a745;">+€${transaction.amount}</span>`;
                    messageText = `<span style="font-size:0.95em;color:#888;">${formattedTimestamp}</span>`;
                } else if (transaction.sender === username) {
                    mainText = `<span style="color:#dc3545;font-weight:600;">Sent</span> <span style="color:#dc3545;">-€${transaction.amount}</span> to <b>${transaction.receiver}</b>`;
                    messageText = `<span style="font-size:0.97em;color:#888;">${formattedTimestamp}</span> <span style="font-style:italic;color:#4e54c8;margin-left:8px;">${transaction.message || 'No message'}</span>`;
                } else {
                    mainText = `<span style="color:#28a745;font-weight:600;">Received</span> <span style="color:#28a745;">+€${transaction.amount}</span> from <b>${transaction.sender}</b>`;
                    messageText = `<span style="font-size:0.97em;color:#888;">${formattedTimestamp}</span> <span style="font-style:italic;color:#4e54c8;margin-left:8px;">${transaction.message || 'No message'}</span>`;
                }

                transactionItem.innerHTML = `<div>${mainText}</div><div style="margin-top:4px;">${messageText}</div>`;
                movementsList.appendChild(transactionItem);
            });
        } catch (error) {
            console.error('Error loading transactions:', error);
            const movementsList = document.getElementById('movements-list');
            movementsList.innerHTML = ''; // Clear existing transactions
            const errorMessage = document.createElement('li');
            errorMessage.textContent = 'Failed to load transactions.';
            errorMessage.style.color = 'red';
            movementsList.appendChild(errorMessage);
        }
    }


    const receiverInput = document.getElementById('receiver-username');
    const suggestionsList = document.getElementById('user-suggestions');
    let selectedUser = null;

    receiverInput.addEventListener('input', async function() {
        const query = receiverInput.value.trim();
        suggestionsList.innerHTML = '';
        selectedUser = null;
        if (query.length < 1) return;
        try {
            const res = await fetch(`/search-users?q=${encodeURIComponent(query)}`);
            if (!res.ok) return;
            const users = await res.json();
            if (!Array.isArray(users)) return;
            users.forEach(user => {
                if (user === sessionStorage.getItem('username')) return; // non mostrare se stessi
                const li = document.createElement('li');
                li.textContent = user;
                li.style.cursor = 'pointer';
                li.style.padding = '10px 14px';
                li.style.margin = '2px 0';
                li.style.borderRadius = '8px';
                li.style.background = '#e9eafc';
                li.style.transition = 'background 0.2s, color 0.2s';
                li.addEventListener('mouseenter', () => {
                    li.style.background = '#4e54c8';
                    li.style.color = '#fff';
                });
                li.addEventListener('mouseleave', () => {
                    li.style.background = '#e9eafc';
                    li.style.color = '#333';
                });
                li.addEventListener('click', function() {
                    receiverInput.value = user;
                    selectedUser = user;
                    suggestionsList.innerHTML = '';
                    receiverInput.style.border = '2px solid #4e54c8';
                    receiverInput.style.background = '#f0f3ff';
                });
                suggestionsList.appendChild(li);
            });
        } catch (e) {
            // Silenzia errori di fetch
        }
    });

});