/**
 * Hatch Homepage — Anime.js Magic Background + Entrance Animations
 *
 * Multi-layered anime.js-powered magical background:
 *   1. Aurora blobs — soft gradient spheres that drift and shift color
 *   2. Rune circles — SVG rings rotating at different speeds
 *   3. Floating orbs — luminous dots with organic drift
 *   4. Energy connections — lines between nearby orbs that pulse
 *   5. Sparkle particles — twinkling dots with staggered opacity
 *   6. Cursor glow — radial glow that follows the mouse
 *
 * Entrance: elements with [data-animate] fade up on scroll.
 * Performance: orb tick throttled to ~30fps, connections drawn every 3rd frame.
 */

(function () {
    'use strict';

    var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    var isMobile = window.innerWidth < 768;

    // ============================================================
    // 1. Aurora Blobs — slow drift + color pulse
    // ============================================================

    function initAurora(animate, random) {
        var blobs = document.querySelectorAll('.magic-aurora');
        blobs.forEach(function (blob, i) {
            animate(blob, {
                translateX: [
                    { to: random(-40, 40), duration: random(8000, 12000) },
                    { to: random(-30, 30), duration: random(8000, 12000) },
                    { to: 0, duration: random(8000, 12000) }
                ],
                translateY: [
                    { to: random(-30, 30), duration: random(10000, 14000) },
                    { to: random(-40, 40), duration: random(10000, 14000) },
                    { to: 0, duration: random(10000, 14000) }
                ],
                scale: [
                    { to: random(0.85, 1.15), duration: random(6000, 10000) },
                    { to: random(0.9, 1.1), duration: random(6000, 10000) },
                    { to: 1, duration: random(6000, 10000) }
                ],
                opacity: [
                    { to: random(0.06, 0.18), duration: random(5000, 8000) },
                    { to: random(0.08, 0.14), duration: random(5000, 8000) },
                    { to: i === 0 ? 0.12 : i === 1 ? 0.1 : 0.08, duration: random(5000, 8000) }
                ],
                loop: true,
                alternate: true,
                ease: 'easeInOutSine'
            });
        });
    }

    // ============================================================
    // 2. Rune Circles — continuous rotation at different speeds
    // ============================================================

    function initRunes(animate, random) {
        var runeData = [
            { sel: '.rune-outer', speed: 120, dir: 1 },
            { sel: '.rune-mid', speed: 80, dir: -1 },
            { sel: '.rune-inner', speed: 60, dir: 1 },
            { sel: '.rune-core', speed: 40, dir: -1 }
        ];

        runeData.forEach(function (r) {
            var el = document.querySelector(r.sel);
            if (!el) return;

            animate(el, {
                opacity: parseFloat(getComputedStyle(el).opacity) || 0.06,
                duration: 2000,
                delay: random(500, 1500),
                ease: 'outQuad'
            });

            animate(el, {
                rotate: r.dir > 0 ? '1turn' : '-1turn',
                duration: r.speed * 1000,
                loop: true,
                ease: 'linear'
            });
        });

        var core = document.querySelector('.rune-core');
        if (core) {
            animate(core, {
                scale: [
                    { to: 1.05, duration: 3000 },
                    { to: 0.95, duration: 3000 },
                    { to: 1, duration: 3000 }
                ],
                loop: true,
                ease: 'easeInOutSine'
            });
        }
    }

    // ============================================================
    // 3. Floating Orbs — luminous dots with organic drift
    // ============================================================

    var orbState = [];

    function createOrbs(count, animate, random) {
        var container = document.getElementById('magic-orbs');
        if (!container) return;

        var colors = [
            'rgba(59, 130, 246, VAR)',
            'rgba(168, 85, 247, VAR)',
            'rgba(6, 182, 212, VAR)',
            'rgba(99, 102, 241, VAR)',
            'rgba(14, 165, 233, VAR)'
        ];

        for (var i = 0; i < count; i++) {
            var size = random(3, 8);
            var opacity = random(0.15, 0.5);
            var colorIdx = Math.min(Math.floor(random(0, colors.length)), colors.length - 1);
            var glowIdx = Math.min(Math.floor(random(0, colors.length)), colors.length - 1);
            var color = colors[colorIdx].replace('VAR', opacity.toFixed(2));
            var glowColor = colors[glowIdx].replace('VAR', (opacity * 0.4).toFixed(2));

            var el = document.createElement('div');
            el.className = 'magic-orb';
            el.style.width = size + 'px';
            el.style.height = size + 'px';
            el.style.background = color;
            el.style.boxShadow = '0 0 ' + (size * 3) + 'px ' + glowColor;
            container.appendChild(el);

            orbState.push({
                x: random(0, 100),
                y: random(0, 100),
                vx: random(-0.015, 0.015),
                vy: random(-0.015, 0.015),
                el: el,
                size: size
            });

            el.style.left = orbState[i].x + '%';
            el.style.top = orbState[i].y + '%';

            animate(el, {
                opacity: [0, 1],
                scale: [0, 1],
                duration: random(800, 1600),
                delay: random(200, 2000),
                ease: 'outExpo'
            });
        }
    }

    // Throttled orb tick — runs at ~30fps for performance
    var lastOrbTick = 0;
    var orbTickInterval = 33; // ~30fps
    var frameCount = 0;

    function tickOrbs(timestamp) {
        requestAnimationFrame(tickOrbs);

        if (timestamp - lastOrbTick < orbTickInterval) return;
        lastOrbTick = timestamp;
        frameCount++;

        orbState.forEach(function (orb) {
            orb.x += orb.vx;
            orb.y += orb.vy;

            if (orb.x < -2 || orb.x > 102) orb.vx *= -1;
            if (orb.y < -2 || orb.y > 102) orb.vy *= -1;

            orb.x = Math.max(-2, Math.min(102, orb.x));
            orb.y = Math.max(-2, Math.min(102, orb.y));

            orb.vx += (Math.random() - 0.5) * 0.0005;
            orb.vy += (Math.random() - 0.5) * 0.0005;

            orb.vx = Math.max(-0.03, Math.min(0.03, orb.vx));
            orb.vy = Math.max(-0.03, Math.min(0.03, orb.vy));

            orb.el.style.left = orb.x + '%';
            orb.el.style.top = orb.y + '%';
        });

        // Draw connections every 3rd frame for performance
        if (frameCount % 3 === 0) {
            drawConnections();
        }
    }

    // ============================================================
    // 4. Energy Connections — lines between nearby orbs
    // ============================================================

    function drawConnections() {
        var svg = document.getElementById('magic-connections');
        if (!svg) return;

        var w = window.innerWidth;
        var h = window.innerHeight;
        var threshold = isMobile ? 15 : 20;

        var lines = '';
        for (var i = 0; i < orbState.length; i++) {
            for (var j = i + 1; j < orbState.length; j++) {
                var dx = orbState[i].x - orbState[j].x;
                var dy = orbState[i].y - orbState[j].y;
                var dist = Math.sqrt(dx * dx + dy * dy);

                if (dist < threshold) {
                    var opacity = (1 - dist / threshold) * 0.12;
                    var x1 = (orbState[i].x / 100) * w;
                    var y1 = (orbState[i].y / 100) * h;
                    var x2 = (orbState[j].x / 100) * w;
                    var y2 = (orbState[j].y / 100) * h;

                    lines += '<line x1="' + x1 + '" y1="' + y1 + '" x2="' + x2 + '" y2="' + y2 + '" stroke="rgba(59,130,246,' + opacity.toFixed(3) + ')" stroke-width="0.5" />';
                }
            }
        }
        svg.innerHTML = lines;
    }

    // ============================================================
    // 5. Sparkle Particles — twinkling dots
    // ============================================================

    function initSparkles(animate, random) {
        var container = document.getElementById('magic-sparkles');
        if (!container) return;

        var count = isMobile ? 15 : 25;
        for (var i = 0; i < count; i++) {
            var dot = document.createElement('div');
            dot.className = 'magic-sparkle';
            dot.style.left = random(5, 95) + '%';
            dot.style.top = random(5, 95) + '%';
            container.appendChild(dot);

            animate(dot, {
                opacity: [
                    { to: random(0, 0.8), duration: random(800, 2000) },
                    { to: 0, duration: random(800, 2000) }
                ],
                scale: [
                    { to: random(0.5, 2), duration: random(800, 2000) },
                    { to: 0.5, duration: random(800, 2000) }
                ],
                delay: random(0, 4000),
                loop: true,
                ease: 'easeInOutSine'
            });
        }
    }

    // ============================================================
    // 6. Cursor Glow — follows mouse
    // ============================================================

    function initCursorGlow(animate) {
        var glow = document.getElementById('magic-cursor-glow');
        if (!glow) return;

        // Skip on touch devices — no cursor to follow
        if ('ontouchstart' in window) return;

        var mouseX = window.innerWidth / 2;
        var mouseY = window.innerHeight / 2;
        var glowX = mouseX;
        var glowY = mouseY;

        document.addEventListener('mousemove', function (e) {
            mouseX = e.clientX;
            mouseY = e.clientY;

            if (glow.style.opacity === '0' || glow.style.opacity === '') {
                animate(glow, { opacity: 1, duration: 600, ease: 'outQuad' });
            }
        });

        function updateGlow() {
            glowX += (mouseX - glowX) * 0.08;
            glowY += (mouseY - glowY) * 0.08;
            glow.style.left = glowX + 'px';
            glow.style.top = glowY + 'px';
            requestAnimationFrame(updateGlow);
        }
        updateGlow();

        document.addEventListener('mouseleave', function () {
            animate(glow, { opacity: 0, duration: 400, ease: 'outQuad' });
        });
    }

    // ============================================================
    // Master background init
    // ============================================================

    function initMagicBackground(animate, createTimeline, stagger, random) {
        initAurora(animate, random);
        initRunes(animate, random);

        var orbCount = isMobile ? 20 : 30;
        createOrbs(orbCount, animate, random);
        requestAnimationFrame(tickOrbs);

        initSparkles(animate, random);
        initCursorGlow(animate);
    }

    // ============================================================
    // Entrance Animations (anime.js for [data-animate])
    // ============================================================

    function initEntranceAnimations(animate, stagger) {
        if (prefersReducedMotion || typeof animate === 'undefined') {
            document.querySelectorAll('[data-animate]').forEach(function (el) {
                el.style.opacity = '1';
                el.style.transform = 'none';
            });
            return;
        }

        // Hero elements animate immediately on load
        var heroElements = document.querySelectorAll('.hero [data-animate]');
        heroElements.forEach(function (el) {
            var delay = parseInt(el.getAttribute('data-delay') || '0', 10);
            animate(el, {
                opacity: [0, 1],
                translateY: [24, 0],
                duration: 700,
                delay: delay + 200,
                ease: 'easeOutQuart'
            });
        });

        // Below-fold elements: Intersection Observer
        var belowFold = document.querySelectorAll('[data-animate]:not(.hero [data-animate])');

        // Fallback: ensure all elements become visible after 1.5s
        // Covers headless/screenshot scenarios where scroll never fires
        setTimeout(function () {
            belowFold.forEach(function (el) {
                if (el.style.opacity === '0' || el.style.opacity === '') {
                    el.style.opacity = '1';
                    el.style.transform = 'none';
                }
            });
        }, 1500);

        if (typeof IntersectionObserver !== 'undefined') {
            var observer = new IntersectionObserver(
                function (entries) {
                    entries.forEach(function (entry) {
                        if (entry.isIntersecting) {
                            var el = entry.target;
                            var delay = parseInt(el.getAttribute('data-delay') || '0', 10);
                            observer.unobserve(el);

                            animate(el, {
                                opacity: [0, 1],
                                translateY: [24, 0],
                                duration: 700,
                                delay: delay,
                                ease: 'easeOutQuart'
                            });
                        }
                    });
                },
                { threshold: 0.1, rootMargin: '0px 0px -40px 0px' }
            );

            belowFold.forEach(function (el) {
                observer.observe(el);
            });
        }
    }

    // ============================================================
    // Terminal glow pulse
    // ============================================================

    function initTerminalGlow(animate) {
        if (prefersReducedMotion || typeof animate === 'undefined') return;

        var terminalEl = document.querySelector('.hero-terminal');
        if (!terminalEl) return;

        setTimeout(function () {
            animate('.hero-terminal', {
                boxShadow: [
                    '0 0 0 0 rgba(59, 130, 246, 0)',
                    '0 0 40px -8px rgba(59, 130, 246, 0.15)',
                    '0 0 0 0 rgba(59, 130, 246, 0)'
                ],
                duration: 2000,
                easing: 'easeInOutQuad'
            });
        }, 1200);
    }

    // ============================================================
    // Boot
    // ============================================================

    function boot(animeLib) {
        var animate = animeLib.animate;
        var createTimeline = animeLib.createTimeline;
        var stagger = animeLib.stagger;
        var random = animeLib.random;

        if (!prefersReducedMotion) {
            initMagicBackground(animate, createTimeline, stagger, random);
        }

        initEntranceAnimations(animate, stagger);
        initTerminalGlow(animate);
    }

    // Defer boot by one frame so anime.js engine is ready
    if (window.anime) {
        requestAnimationFrame(function () { boot(window.anime); });
    } else {
        var tries = 0;
        var poll = setInterval(function () {
            tries++;
            if (window.anime) {
                clearInterval(poll);
                requestAnimationFrame(function () { boot(window.anime); });
            }
            if (tries > 50) clearInterval(poll);
        }, 100);
    }

})();
