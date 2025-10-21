# Future Enhancements

This document tracks potential features and improvements to be considered for future implementation.

## Deferred Phase 3 Features

### 1. SVG Timeline for "Our Story" Section
**Estimated Effort:** 8 hours
**Wow Factor:** 8/10

Transform the narrative text into an interactive timeline with animated SVG path drawing and scroll-triggered reveals.

**Key Features:**
- Vertical SVG path that draws as user scrolls (stroke-dashoffset animation)
- Timeline nodes (circles) appear at key moments:
  - 2018: First met at brunch in Southie
  - Halloween 2018: First date at Fenway Johnnie's
  - 2024: Engagement at Eastern Standard
- Story text fades in beside each node with directional slides (alternating left/right)
- Dates and photos attached to timeline points
- Subtle particle effects (hearts, sparkles) that float along the timeline

**Technical Approach:**
- Calculate SVG path length using `path.getTotalLength()`
- Use Intersection Observer to track scroll position
- Map scroll progress to stroke-dashoffset
- Trigger text animations when dots enter viewport
- CSS custom properties for dynamic theming

**Why It's Valuable:**
Demonstrates SVG mastery, scroll-driven animation choreography, and storytelling through motion. Creates emotional connection while showcasing technical skill.

---

### 2. Interactive Guest Book Section
**Estimated Effort:** 10 hours
**Wow Factor:** 8/10

Interactive section where visitors can leave messages, with beautiful card-based UI and optional moderation.

**Key Features:**
- **Frontend:**
  - Form with character counter and emoji picker
  - Card-based message display (masonry layout)
  - Messages appear with staggered fade-in animation
  - Hover effects: Cards slightly lift and rotate
  - Heart icon to "like" messages (increment counter)
  - Real-time updates (WebSocket or polling)
  - Optional: Pin favorite messages to top

- **Backend:**
  - Lambda function for message storage
  - DynamoDB or S3 for persistence
  - Moderation panel to approve/hide messages
  - Rate limiting (1 message per IP per hour)
  - Spam prevention

**Visual Design:**
- Each message card: White background, soft shadow, rounded corners
- Author name in script font, message in sans-serif
- Timestamp in subtle gray
- Optimistic UI updates (show message immediately, sync with backend)

**Why It's Valuable:**
Demonstrates full-stack capability, real-time data handling, form validation, and community engagement. Major portfolio piece showing ability to build interactive features, not just static pages.

---

## Other Potential Enhancements

### Performance Optimizations
- Convert images to WebP with JPEG fallback
- Implement responsive images (srcset, sizes)
- Add service worker for offline support
- Code splitting for route-specific JavaScript

### Accessibility Improvements
- WCAG 2.1 AAA compliance audit
- Enhanced keyboard navigation
- Screen reader testing and improvements
- Focus trap in modals

### Advanced Gallery Features
- Category filtering for photos
- Favorites feature with localStorage
- Photo comparison slider (before/after style)
- Search functionality

### Design Enhancements
- Dark mode toggle
- Parallax scroll effects
- Animated gradient backgrounds
- Micro-interactions on all interactive elements
- Countdown timer to wedding date

---

## Notes
Features are organized by complexity and impact. Revisit this list after Phase 3 implementation to prioritize next steps.
