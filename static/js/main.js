// Main JavaScript File

// Global variables
const BASE_URL = window.location.origin;
let currentUser = null;

// DOM Elements
const menuIcon = document.getElementById('menuIcon');
const closeMenu = document.getElementById('closeMenu');
const navLinks = document.getElementById('navLinks');
const bookNowBtn = document.getElementById('bookNowBtn');
const quickBookingForm = document.getElementById('quickBookingForm');

// Navigation functionality
menuIcon.addEventListener('click', () => {
  navLinks.classList.add('active');
});

closeMenu.addEventListener('click', () => {
  navLinks.classList.remove('active');
});

// Close navigation when clicking a link on mobile
document.querySelectorAll('.nav-links a').forEach(link => {
  link.addEventListener('click', () => {
    if (window.innerWidth <= 768) {
      navLinks.classList.remove('active');
    }
  });
});

// Navbar scroll effect
window.addEventListener('scroll', () => {
  const navbar = document.querySelector('.navbar');
  if (window.scrollY > 100) {
    navbar.style.padding = '5px 0';
    navbar.style.backgroundColor = 'rgba(255, 255, 255, 0.95)';
    navbar.style.boxShadow = '0 4px 6px rgba(0, 0, 0, 0.1)';
  } else {
    navbar.style.padding = '8px 0';
    navbar.style.backgroundColor = 'var(--neutral-100)';
    navbar.style.boxShadow = 'var(--shadow-sm)';
  }
});

// Set minimum date for appointment date input to today
if (document.getElementById('appointmentDate')) {
  const today = new Date().toISOString().split('T')[0];
  document.getElementById('appointmentDate').min = today;
}

// Quick Booking Form Submission
if (quickBookingForm) {
  quickBookingForm.addEventListener('submit', (e) => {
    e.preventDefault();
    
    const speciality = document.getElementById('speciality').value;
    const location = document.getElementById('location').value;
    const date = document.getElementById('appointmentDate').value;
    
    // Simulate form submission - in a real app, this would redirect to a search results page
    alert(`Searching for ${speciality} doctors in ${location} for ${date}`);
    
    // For demo purposes, scroll to doctors section
    document.getElementById('doctors').scrollIntoView({ behavior: 'smooth' });
  });
}

// Book Now button functionality
if (bookNowBtn) {
  bookNowBtn.addEventListener('click', () => {
    // Scroll to quick booking widget
    document.querySelector('.booking-widget').scrollIntoView({ behavior: 'smooth' });
  });
}

// Learn More button functionality
const learnMoreBtn = document.getElementById('learnMoreBtn');
if (learnMoreBtn) {
  learnMoreBtn.addEventListener('click', () => {
    // Scroll to services section
    document.getElementById('services').scrollIntoView({ behavior: 'smooth' });
  });
}

// Form validation function
function validateForm(form) {
  let isValid = true;
  const inputs = form.querySelectorAll('input[required], select[required], textarea[required]');
  
  inputs.forEach(input => {
    if (!input.value.trim()) {
      isValid = false;
      input.style.borderColor = 'var(--error)';
      
      // Add error message if it doesn't exist
      let errorElement = input.nextElementSibling;
      if (!errorElement || !errorElement.classList.contains('error-message')) {
        errorElement = document.createElement('div');
        errorElement.classList.add('error-message');
        errorElement.style.color = 'var(--error)';
        errorElement.style.fontSize = '0.8rem';
        errorElement.style.marginTop = '4px';
        input.parentNode.insertBefore(errorElement, input.nextSibling);
      }
      errorElement.textContent = 'This field is required';
    } else {
      input.style.borderColor = 'var(--neutral-400)';
      
      // Remove error message if it exists
      const errorElement = input.nextElementSibling;
      if (errorElement && errorElement.classList.contains('error-message')) {
        errorElement.remove();
      }
    }
  });
  
  return isValid;
}

// Contact form submission
const contactForm = document.getElementById('contactForm');
if (contactForm) {
  contactForm.addEventListener('submit', (e) => {
    e.preventDefault();
    
    if (validateForm(contactForm)) {
      // Get form data
      const name = document.getElementById('name').value;
      const email = document.getElementById('email').value;
      const subject = document.getElementById('subject').value;
      const message = document.getElementById('message').value;
      
      // In a real app, this would send the data to a server
      alert(`Thank you, ${name}! Your message has been sent.`);
      contactForm.reset();
    }
  });
}

// Newsletter form submission
const newsletterForm = document.getElementById('newsletterForm');
if (newsletterForm) {
  newsletterForm.addEventListener('submit', (e) => {
    e.preventDefault();
    
    const emailInput = newsletterForm.querySelector('input[type="email"]');
    const email = emailInput.value;
    
    if (email) {
      // In a real app, this would send the email to a server
      alert(`Thank you for subscribing with ${email}!`);
      newsletterForm.reset();
    } else {
      emailInput.style.borderColor = 'var(--error)';
    }
  });
}

// Initialize date picker for today's date by default
document.addEventListener('DOMContentLoaded', () => {
  if (document.getElementById('appointmentDate')) {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('appointmentDate').value = today;
  }
});