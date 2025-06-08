document.addEventListener('DOMContentLoaded', () => {
  const chatToggle = document.querySelector('.chat-toggle');
  const chatModal = document.querySelector('.chat-modal');
  const closeChat = document.querySelector('.close-chat');
  const modeButtons = document.querySelectorAll('.mode-btn');
  const chatInput = document.querySelector('.chat-input input');
  const sendBtn = document.querySelector('.send-btn');
  const chatMessages = document.querySelector('.chat-messages');
  const diseaseForm = document.querySelector('.disease-form');
  const generalInput = document.querySelector('.chat-input');

  let currentMode = 'general';

  chatToggle.addEventListener('click', () => {
    chatModal.classList.toggle('active');
  });

  closeChat.addEventListener('click', () => {
    chatModal.classList.remove('active');
  });

  modeButtons.forEach(btn => {
    btn.addEventListener('click', () => {
      modeButtons.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      currentMode = btn.dataset.mode;
      
      if (currentMode === 'disease') {
        diseaseForm.style.display = 'block';
        generalInput.style.display = 'none';
      } else {
        diseaseForm.style.display = 'none';
        generalInput.style.display = 'flex';
      }
    });
  });

  document.getElementById('diseaseForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());
    
    addMessage('user', `Age: ${data.age}, Gender: ${data.gender}, Symptoms: ${data.symptoms}, Medical History: ${data.medical_history}`);
    document.getElementById('general_section').click();
    try {
      const response = await fetch('/predict-disease', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data)
      });
      
      const result = await response.json();
      addMessage('bot', result.response);
    } catch (error) {
      console.error('Error:', error);
      addMessage('bot', 'Sorry, there was an error processing your request.');
    }
  });

  sendBtn.addEventListener('click', async () => {
    const message = chatInput.value.trim();
    if (!message) return;
    
    addMessage('user', message);
    chatInput.value = '';
    
    try {
      const response = await fetch('/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message })
      });
      
      const result = await response.json();
      addMessage('bot', result.response);
    } catch (error) {
      console.error('Error:', error);
      addMessage('bot', 'Sorry, there was an error processing your query.');
    }
  });

  function addMessage(sender, text) {
    const messageDiv = document.createElement('div');
    messageDiv.classList.add('message', `${sender}-message`);
    messageDiv.textContent = text;
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }
});