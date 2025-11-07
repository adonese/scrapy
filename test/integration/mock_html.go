package integration

// MockBayutHTML contains a mock HTML response from Bayut
const MockBayutHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Properties for Rent in Dubai - Bayut</title>
</head>
<body>
    <article data-testid="property-card">
        <a href="/property/details-12345" title="Spacious 1BR Apartment in Dubai Marina">
            <h2>Spacious 1BR Apartment in Dubai Marina</h2>
        </a>
        <span aria-label="Price">AED 85,000/year</span>
        <div aria-label="Location">Dubai Marina, Dubai</div>
        <span aria-label="Bedrooms">1 Bedroom</span>
    </article>

    <article data-testid="property-card">
        <a href="/property/details-12346" title="Modern Studio in Business Bay">
            <h2>Modern Studio in Business Bay</h2>
        </a>
        <span aria-label="Price">AED 55,000/year</span>
        <div aria-label="Location">Business Bay, Dubai</div>
        <span aria-label="Bedrooms">Studio</span>
    </article>

    <article data-testid="property-card">
        <a href="/property/details-12347" title="Luxury 2BR in Downtown Dubai">
            <h2>Luxury 2BR in Downtown Dubai</h2>
        </a>
        <span aria-label="Price">AED 120,000/year</span>
        <div aria-label="Location">Downtown Dubai, Dubai</div>
        <span aria-label="Bedrooms">2 BR</span>
    </article>

    <article data-testid="property-card">
        <a href="/property/details-12348" title="Cozy 1BR in Al Nahda">
            <h2>Cozy 1BR in Al Nahda</h2>
        </a>
        <span aria-label="Price">AED 45,000/year</span>
        <div aria-label="Location">Al Nahda, Sharjah</div>
        <span aria-label="Bedrooms">1 Bedroom</span>
    </article>

    <article data-testid="property-card">
        <a href="/property/details-12349" title="Spacious 3BR Villa">
            <h2>Spacious 3BR Villa</h2>
        </a>
        <span aria-label="Price">AED 180,000/year</span>
        <div aria-label="Location">Al Reem Island, Abu Dhabi</div>
        <span aria-label="Bedrooms">3 BR</span>
    </article>
</body>
</html>
`

// MockBayutEmptyHTML contains an empty HTML response from Bayut
const MockBayutEmptyHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Properties for Rent - Bayut</title>
</head>
<body>
    <div class="no-results">No properties found</div>
</body>
</html>
`

// MockBayutMalformedHTML contains malformed HTML
const MockBayutMalformedHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Properties for Rent - Bayut</title>
</head>
<body>
    <article data-testid="property-card">
        <a href="/property/details-12345">
            <h2>Missing Price</h2>
        </a>
        <div aria-label="Location">Dubai Marina, Dubai</div>
    </article>
</body>
</html>
`

// MockDubizzleHTML contains a mock HTML response from Dubizzle
const MockDubizzleHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Property for Rent in Dubai - Dubizzle</title>
</head>
<body>
    <li data-testid="listing-item">
        <a href="/property-for-rent/apartments-flats/dubai/dubai-marina/12345">
            <h2>Spacious 1BR in Marina</h2>
        </a>
        <span data-testid="listing-price">AED 75,000</span>
        <span data-testid="listing-location">Dubai Marina, Dubai</span>
        <span data-testid="bedrooms">1 Bedroom</span>
        <span data-testid="bathrooms">1 Bathroom</span>
    </li>

    <li data-testid="listing-item">
        <a href="/property-for-rent/apartments-flats/dubai/business-bay/12346">
            <h2>Affordable Studio</h2>
        </a>
        <span data-testid="listing-price">AED 50,000</span>
        <span data-testid="listing-location">Business Bay - Dubai</span>
        <span data-testid="bedrooms">Studio</span>
        <span data-testid="bathrooms">1 Bath</span>
    </li>

    <li data-testid="listing-item">
        <a href="/property-for-rent/apartments-flats/dubai/jlt/12347">
            <h2>Modern 2BR Apartment</h2>
        </a>
        <span data-testid="listing-price">95,000 AED/year</span>
        <span data-testid="listing-location">Jumeirah Lake Towers | Dubai</span>
        <span data-testid="bedrooms">2 BR</span>
        <span data-testid="bathrooms">2 Baths</span>
    </li>

    <li data-testid="listing-item">
        <a href="/property-for-rent/apartments-flats/sharjah/al-nahda/12348">
            <h2>Budget 1BR</h2>
        </a>
        <span data-testid="listing-price">40,000 Dhs</span>
        <span data-testid="listing-location">Al Nahda, Sharjah</span>
        <span data-testid="bedrooms">1 Bed</span>
        <span data-testid="bathrooms">1</span>
    </li>

    <li data-testid="listing-item">
        <a href="/property-for-rent/bedspace/dubai/bur-dubai/12349">
            <h2>Bed Space in Shared Room</h2>
        </a>
        <span data-testid="listing-price">AED 800/month</span>
        <span data-testid="listing-location">Bur Dubai, Dubai</span>
    </li>
</body>
</html>
`

// MockDubizzleEmptyHTML contains an empty HTML response from Dubizzle
const MockDubizzleEmptyHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Property for Rent - Dubizzle</title>
</head>
<body>
    <div class="no-listings">
        <p>No listings found matching your criteria</p>
    </div>
</body>
</html>
`

// MockDubizzleBotDetectedHTML contains a bot detection page
const MockDubizzleBotDetectedHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Access Denied</title>
</head>
<body>
    <div class="incapsula">
        <h1>Access Denied</h1>
        <p>Incapsula incident ID: 12345-67890</p>
        <p>Your access has been blocked due to suspicious activity.</p>
    </div>
</body>
</html>
`

// MockDubizzleCloudflareHTML contains a Cloudflare challenge page
const MockDubizzleCloudflareHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Just a moment...</title>
</head>
<body>
    <div class="cloudflare">
        <h1>Checking your browser before accessing the website</h1>
        <p>This process is automatic. Your browser will redirect to your requested content shortly.</p>
        <p>Ray ID: abc123def456</p>
    </div>
</body>
</html>
`

// MockDubizzleMalformedHTML contains malformed HTML
const MockDubizzleMalformedHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Property for Rent - Dubizzle</title>
</head>
<body>
    <li data-testid="listing-item">
        <a href="/property-for-rent/apartments-flats/dubai/marina/12345">
            <h2>Incomplete Listing</h2>
        </a>
        <!-- Missing price -->
        <span data-testid="listing-location">Dubai Marina</span>
    </li>
</body>
</html>
`
