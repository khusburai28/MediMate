// Modal Functionality for Login and Signup

// DOM Elements
const loginModal = document.getElementById('loginModal');
const signupModal = document.getElementById('signupModal');
const loginBtn = document.getElementById('loginBtn');
const signupBtn = document.getElementById('signupBtn');
const closeLoginModal = document.getElementById('closeLoginModal');
const closeSignupModal = document.getElementById('closeSignupModal');
const switchToSignup = document.getElementById('switchToSignup');
const switchToLogin = document.getElementById('switchToLogin');
const loginForm = document.getElementById('loginForm');
const signupForm = document.getElementById('signupForm');
const userTypeTabs = document.querySelectorAll('.user-type-tabs .tab-btn');

// Show login modal
if (loginBtn) {
  loginBtn.addEventListener('click', () => {
    loginModal.style.display = 'flex';
  });
}

// Show signup modal
if (signupBtn) {
  signupBtn.addEventListener('click', () => {
    signupModal.style.display = 'flex';
  });
}

// Close login modal
if (closeLoginModal) {
  closeLoginModal.addEventListener('click', () => {
    loginModal.style.display = 'none';
  });
}

// Close signup modal
if (closeSignupModal) {
  closeSignupModal.addEventListener('click', () => {
    signupModal.style.display = 'none';
  });
}

// Switch to signup from login
if (switchToSignup) {
  switchToSignup.addEventListener('click', (e) => {
    e.preventDefault();
    loginModal.style.display = 'none';
    signupModal.style.display = 'flex';
  });
}

// Switch to login from signup
if (switchToLogin) {
  switchToLogin.addEventListener('click', (e) => {
    e.preventDefault();
    signupModal.style.display = 'none';
    loginModal.style.display = 'flex';
  });
}

// Close modals when clicking outside of modal content
window.addEventListener('click', (e) => {
  if (e.target === loginModal) {
    loginModal.style.display = 'none';
  }
  if (e.target === signupModal) {
    signupModal.style.display = 'none';
  }
});

// Handle user type tabs
userTypeTabs.forEach(tab => {
  tab.addEventListener('click', () => {
    // Remove active class from all tabs
    userTypeTabs.forEach(t => t.classList.remove('active'));
    // Add active class to clicked tab
    tab.classList.add('active');
    
    // Additional logic to update form fields based on user type
    const userType = tab.getAttribute('data-type');
    updateFormFields(userType);
  });
});

// Update form fields based on user type
function updateFormFields(userType) {
  if (userType === 'doctor' && signupForm) {
    const specializationField = document.querySelector('.specialization-field');
    if (specializationField) {
      specializationField.style.display = 'block';
    } else {
      const formGroups = signupForm.querySelectorAll('.form-group');
      const lastFormGroup = formGroups[formGroups.length - 2];
      
      const specializationGroup = document.createElement('div');
      specializationGroup.className = 'form-group specialization-field';
      specializationGroup.innerHTML = `
        <label for="specialization">Specialization</label>
        <select id="specialization" required>
          <option value="" disabled selected>Select specialization</option>
          <option value="cardiology">Cardiology</option>
          <option value="dermatology">Dermatology</option>
          <option value="neurology">Neurology</option>
          <option value="orthopedics">Orthopedics</option>
          <option value="pediatrics">Pediatrics</option>
          <option value="psychiatry">Psychiatry</option>
        </select>
      `;
      
      if (lastFormGroup) {
        signupForm.insertBefore(specializationGroup, lastFormGroup);
      }
    }
  } else if (userType !== 'doctor' && signupForm) {
    const specializationField = document.querySelector('.specialization-field');
    if (specializationField) {
      specializationField.style.display = 'none';
    }
  }
}

// Login form submission
// if (loginForm) {
//   loginForm.addEventListener('submit', (e) => {
//     e.preventDefault();
    
//     const email = document.getElementById('loginEmail').value;
//     const password = document.getElementById('loginPassword').value;
//     const userType = document.querySelector('.user-type-tabs .tab-btn.active').getAttribute('data-type');
    
//     // Simulate login and redirect to appropriate dashboard
//     simulateLogin(email, userType);
//   });
// }

// // Signup form submission
// if (signupForm) {
//   signupForm.addEventListener('submit', (e) => {
//     e.preventDefault();
    
//     const name = document.getElementById('signupName').value;
//     const email = document.getElementById('signupEmail').value;
//     const phone = document.getElementById('signupPhone').value;
//     const password = document.getElementById('signupPassword').value;
//     const confirmPassword = document.getElementById('confirmPassword').value;
//     const userType = document.querySelector('.user-type-tabs .tab-btn.active').getAttribute('data-type');
    
//     if (password !== confirmPassword) {
//       alert('Passwords do not match');
//       return;
//     }
    
//     // Simulate registration and redirect to appropriate dashboard
//     simulateRegistration(name, email, userType);
//   });
// }

// // Simulate login
// function simulateLogin(email, userType) {
//   currentUser = {
//     email: email,
//     name: email.split('@')[0],
//     userType: userType
//   };
  
//   loginModal.style.display = 'none';
  
//   // Redirect based on user type
//   if (userType === 'doctor') {
//     window.location.href = 'pages/doctor-dashboard.html';
//   } else if (userType === 'patient') {
//     window.location.href = 'pages/patient-dashboard.html';
//   } else if (userType === 'admin') {
//     alert('Admin dashboard is under development');
//   }
// }

// // Simulate registration
// function simulateRegistration(name, email, userType) {
//   currentUser = {
//     name: name,
//     email: email,
//     userType: userType
//   };
  
//   signupModal.style.display = 'none';
  
//   // Redirect based on user type
//   if (userType === 'doctor') {
//     window.location.href = 'pages/doctor-dashboard.html';
//   } else if (userType === 'patient') {
//     window.location.href = 'pages/patient-dashboard.html';
//   }
// }