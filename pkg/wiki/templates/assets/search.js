document.addEventListener('DOMContentLoaded', function() {
    // Handle sidebar search
    const sidebarSearch = document.getElementById('search');
    if (sidebarSearch) {
        sidebarSearch.addEventListener('input', function(e) {
            const query = e.target.value.toLowerCase();
            const navItems = document.querySelectorAll('.nav-items li a');
            
            navItems.forEach(item => {
                const text = item.textContent.toLowerCase();
                const li = item.parentElement;
                if (text.includes(query)) {
                    li.style.display = '';
                    // Show parent group if it's a nested item
                    const parentGroup = li.parentElement.parentElement;
                    if (parentGroup.classList.contains('nav-items')) {
                        parentGroup.style.display = '';
                    }
                } else {
                    li.style.display = 'none';
                }
            });
        });
    }

    // Handle main search page
    const searchPage = document.getElementById('search-page');
    const searchButton = document.getElementById('search-button');
    const searchResults = document.getElementById('search-results');

    if (searchPage && searchButton && searchResults) {
        searchButton.addEventListener('click', performSearch);
        searchPage.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                performSearch();
            }
        });
    }

    async function performSearch() {
        const query = searchPage.value;
        if (!query) return;

        try {
            const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
            const results = await response.json();
            
            displaySearchResults(results);
        } catch (error) {
            console.error('Search failed:', error);
            searchResults.innerHTML = '<p class="error">Search failed. Please try again.</p>';
        }
    }

    function displaySearchResults(results) {
        if (!results || results.length === 0) {
            searchResults.innerHTML = '<p>No results found.</p>';
            return;
        }

        const html = results.map(result => `
            <div class="search-result">
                <h3><a href="${result.url}">${result.title}</a></h3>
                <p>${result.excerpt}</p>
                <p class="search-meta">Score: ${result.score.toFixed(2)}</p>
            </div>
        `).join('');

        searchResults.innerHTML = html;
    }
});
