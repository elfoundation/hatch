/**
 * Hatch Homepage v2 — Three.js particle network + anime.js entrance animations
 *
 * Particle network: nodes float gently, connected by proximity lines.
 * Represents HTTP requests flowing through the network.
 * Entrance: elements with [data-animate] fade up on scroll.
 */

(function () {
    'use strict';

    // ============================================================
    // Three.js Particle Network
    // ============================================================

    const canvas = document.getElementById('particle-canvas');
    if (!canvas) return;

    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    if (!prefersReducedMotion && typeof THREE !== 'undefined') {
        initParticleNetwork();
    }

    function initParticleNetwork() {
        const scene = new THREE.Scene();
        const camera = new THREE.PerspectiveCamera(
            60,
            window.innerWidth / window.innerHeight,
            0.1,
            1000
        );
        camera.position.z = 50;

        const renderer = new THREE.WebGLRenderer({
            canvas: canvas,
            alpha: true,
            antialias: false,
        });
        renderer.setSize(window.innerWidth, window.innerHeight);
        renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

        // Particle parameters
        const PARTICLE_COUNT = 80;
        const CONNECTION_DISTANCE = 12;
        const PARTICLE_SIZE = 0.15;

        // Create particles
        const particlesGeometry = new THREE.BufferGeometry();
        const positions = new Float32Array(PARTICLE_COUNT * 3);
        const velocities = [];
        const opacities = new Float32Array(PARTICLE_COUNT);

        for (let i = 0; i < PARTICLE_COUNT; i++) {
            positions[i * 3] = (Math.random() - 0.5) * 100;
            positions[i * 3 + 1] = (Math.random() - 0.5) * 60;
            positions[i * 3 + 2] = (Math.random() - 0.5) * 30;

            velocities.push({
                x: (Math.random() - 0.5) * 0.02,
                y: (Math.random() - 0.5) * 0.02,
                z: (Math.random() - 0.5) * 0.01,
            });

            opacities[i] = 0.3 + Math.random() * 0.5;
        }

        particlesGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

        const particlesMaterial = new THREE.PointsMaterial({
            color: 0x3b82f6,
            size: PARTICLE_SIZE,
            transparent: true,
            opacity: 0.6,
            sizeAttenuation: true,
            blending: THREE.AdditiveBlending,
            depthWrite: false,
        });

        const particles = new THREE.Points(particlesGeometry, particlesMaterial);
        scene.add(particles);

        // Connection lines
        const lineMaterial = new THREE.LineBasicMaterial({
            color: 0x3b82f6,
            transparent: true,
            opacity: 0.08,
            blending: THREE.AdditiveBlending,
            depthWrite: false,
        });

        const linesGroup = new THREE.Group();
        scene.add(linesGroup);

        // Mouse interaction
        let mouseX = 0;
        let mouseY = 0;
        document.addEventListener('mousemove', (e) => {
            mouseX = (e.clientX / window.innerWidth - 0.5) * 2;
            mouseY = -(e.clientY / window.innerHeight - 0.5) * 2;
        });

        // Resize
        function onResize() {
            camera.aspect = window.innerWidth / window.innerHeight;
            camera.updateProjectionMatrix();
            renderer.setSize(window.innerWidth, window.innerHeight);
        }
        window.addEventListener('resize', onResize);

        // Animate
        function animate() {
            requestAnimationFrame(animate);

            const posArray = particlesGeometry.attributes.position.array;

            // Update particle positions
            for (let i = 0; i < PARTICLE_COUNT; i++) {
                posArray[i * 3] += velocities[i].x;
                posArray[i * 3 + 1] += velocities[i].y;
                posArray[i * 3 + 2] += velocities[i].z;

                // Wrap around edges
                if (posArray[i * 3] > 55) posArray[i * 3] = -55;
                if (posArray[i * 3] < -55) posArray[i * 3] = 55;
                if (posArray[i * 3 + 1] > 35) posArray[i * 3 + 1] = -35;
                if (posArray[i * 3 + 1] < -35) posArray[i * 3 + 1] = 35;
                if (posArray[i * 3 + 2] > 20) posArray[i * 3 + 2] = -20;
                if (posArray[i * 3 + 2] < -20) posArray[i * 3 + 2] = 20;
            }

            particlesGeometry.attributes.position.needsUpdate = true;

            // Update connection lines
            while (linesGroup.children.length > 0) {
                const child = linesGroup.children[0];
                child.geometry.dispose();
                child.material.dispose();
                linesGroup.remove(child);
            }

            for (let i = 0; i < PARTICLE_COUNT; i++) {
                for (let j = i + 1; j < PARTICLE_COUNT; j++) {
                    const dx = posArray[i * 3] - posArray[j * 3];
                    const dy = posArray[i * 3 + 1] - posArray[j * 3 + 1];
                    const dz = posArray[i * 3 + 2] - posArray[j * 3 + 2];
                    const dist = Math.sqrt(dx * dx + dy * dy + dz * dz);

                    if (dist < CONNECTION_DISTANCE) {
                        const lineGeometry = new THREE.BufferGeometry();
                        const linePositions = new Float32Array([
                            posArray[i * 3], posArray[i * 3 + 1], posArray[i * 3 + 2],
                            posArray[j * 3], posArray[j * 3 + 1], posArray[j * 3 + 2],
                        ]);
                        lineGeometry.setAttribute('position', new THREE.BufferAttribute(linePositions, 3));

                        const lineMat = lineMaterial.clone();
                        lineMat.opacity = 0.06 * (1 - dist / CONNECTION_DISTANCE);
                        linesGroup.add(new THREE.Line(lineGeometry, lineMat));
                    }
                }
            }

            // Subtle camera follow mouse
            camera.position.x += (mouseX * 3 - camera.position.x) * 0.02;
            camera.position.y += (mouseY * 2 - camera.position.y) * 0.02;
            camera.lookAt(scene.position);

            renderer.render(scene, camera);
        }

        animate();
    }

    // ============================================================
    // Anime.js Entrance Animations
    // ============================================================

    if (prefersReducedMotion || typeof anime === 'undefined') {
        // Make everything visible immediately
        document.querySelectorAll('[data-animate]').forEach(function (el) {
            el.style.opacity = '1';
            el.style.transform = 'none';
        });
        return;
    }

    // Hero elements animate immediately on load (above the fold)
    var heroElements = document.querySelectorAll('.hero [data-animate]');
    heroElements.forEach(function (el) {
        var delay = parseInt(el.getAttribute('data-delay') || '0', 10);
        anime({
            targets: el,
            opacity: [0, 1],
            translateY: [24, 0],
            duration: 700,
            delay: delay + 200, // small base offset for page load
            easing: 'easeOutQuart',
        });
    });

    // Below-fold elements: Intersection Observer with immediate fallback
    var belowFold = document.querySelectorAll('[data-animate]:not(.hero [data-animate])');

    // Fallback: ensure all elements become visible after 1.5s max
    // This covers headless/screenshot scenarios where scroll never fires
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

                        anime({
                            targets: el,
                            opacity: [0, 1],
                            translateY: [24, 0],
                            duration: 700,
                            delay: delay,
                            easing: 'easeOutQuart',
                        });
                    }
                });
            },
            {
                threshold: 0.1,
                rootMargin: '0px 0px -40px 0px',
            }
        );

        belowFold.forEach(function (el) {
            observer.observe(el);
        });
    }

    // ============================================================
    // Terminal glow pulse (subtle)
    // ============================================================

    var terminalEl = document.querySelector('.hero-terminal');
    if (terminalEl && !prefersReducedMotion && typeof anime !== 'undefined') {
        setTimeout(function () {
            anime({
                targets: '.hero-terminal',
                boxShadow: [
                    '0 0 0 0 rgba(59, 130, 246, 0)',
                    '0 0 40px -8px rgba(59, 130, 246, 0.15)',
                    '0 0 0 0 rgba(59, 130, 246, 0)',
                ],
                duration: 2000,
                easing: 'easeInOutQuad',
            });
        }, 1200);
    }
})();
