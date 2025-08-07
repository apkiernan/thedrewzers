// First View Handler - Uses localStorage instead of cookies
(function() {
    // Check if user has seen the first view
    const hasSeenFirstView = localStorage.getItem('first_view_seen') === 'true';
    
    if (!hasSeenFirstView) {
        // Create and inject the first view overlay
        const firstViewHTML = `
<div id="first-view-overlay" class="first-view-overlay">
    <div class="slideshow-container">
        <div class="floating-hearts">
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
            <div class="heart">♥</div>
        </div>
        
        <div class="text-line line-1">You've waited long enough</div>
        <div class="text-line line-2">We've been asked many times</div>
        <div class="text-line line-3">SO many times</div>
        <div class="text-line line-4">The time has come</div>
        <div class="text-line line-5">May 30, 2026</div>
        <div class="text-line line-6">The Tower, 101 Arlington Street, Boston, MA</div>
        <div class="text-line line-7">The Kiernan/Smith Wedding</div>
        
        <div class="cta-section">
            <a href="#view-details" class="cta-button">View Details</a>
        </div>
    </div>
</div>`;
        
        // Add the overlay to the page
        document.body.insertAdjacentHTML('afterbegin', firstViewHTML);
        
        // Add event listener for the CTA button
        setTimeout(() => {
            const viewDetailsButton = document.querySelector('.cta-button');
            if (viewDetailsButton) {
                viewDetailsButton.addEventListener('click', function(e) {
                    e.preventDefault();
                    
                    // Fade out the overlay
                    const overlay = document.getElementById('first-view-overlay');
                    overlay.classList.add('fade-out');
                    
                    // Mark as seen
                    localStorage.setItem('first_view_seen', 'true');
                    
                    // Remove after animation
                    setTimeout(() => {
                        overlay.remove();
                    }, 1000);
                });
            }
        }, 100);
    }
})();