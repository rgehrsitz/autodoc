// Search functionality
document.addEventListener('DOMContentLoaded', function() {
    const searchInput = document.getElementById('search');
    if (!searchInput) return;

    searchInput.addEventListener('input', function(e) {
        const query = e.target.value.toLowerCase();
        const components = document.querySelectorAll('.component');
        
        components.forEach(component => {
            const title = component.querySelector('.component-title').textContent.toLowerCase();
            const description = component.querySelector('.component-description').textContent.toLowerCase();
            
            if (title.includes(query) || description.includes(query)) {
                component.style.display = '';
            } else {
                component.style.display = 'none';
            }
        });
    });
});
