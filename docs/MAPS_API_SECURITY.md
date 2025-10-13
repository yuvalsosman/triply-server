# Google Maps API Security Implementation

## Overview

The Google Maps API key is now managed server-side to provide a seamless user experience while maintaining security. Users no longer need to provide their own API keys.

## Architecture

### Backend (Go/Fiber)

1. **Configuration (`internal/config/config.go`)**
   - Added `MapsConfig` struct with `APIKey` field
   - Loads `GOOGLE_MAPS_API_KEY` from environment variables

2. **Handler (`internal/handlers/maps_handler.go`)**
   - `NewMapsHandler`: Constructor accepting API key
   - `GetMapConfig`: Endpoint that returns the API key to clients

3. **Routes (`cmd/server/main.go`)**
   - Added route: `GET /api/maps/config`
   - Public endpoint (no authentication required)
   - Returns: `{ "apiKey": "AIzaSy..." }`

### Frontend (Next.js/React)

1. **Service (`lib/services/apiKeyService.ts`)**
   - Added `fetchFromServer()` method
   - Fetches API key from backend on demand
   - Legacy localStorage methods kept for backwards compatibility

2. **Components**
   - `TripEditor.tsx`: Updated to fetch key from server
   - `app/explore/[tripId]/page.tsx`: Updated to fetch key from server
   - API key modals commented out (no longer needed)

## Security Layers

### 1. HTTP Referrer Restrictions (Primary)
Configure in Google Cloud Console:
- **Application restrictions**: "HTTP referrers (web sites)"
- **Allowed referrers**: 
  - `http://localhost:3001/*` (development)
  - `https://yourdomain.com/*` (production)

This prevents the API key from working on any other domain, even if exposed.

### 2. API Restrictions
Configure in Google Cloud Console:
- **API restrictions**: Limit to specific APIs
- **Enabled APIs**:
  - Maps JavaScript API
  - Places API

This prevents the key from being used for other Google services.

### 3. Usage Quotas
Configure in Google Cloud Console:
- Set daily quotas to prevent cost overruns
- Monitor usage regularly
- Set up billing alerts

## Benefits

1. **User Experience**: Users get maps immediately without configuration
2. **Security**: Key protected by HTTP referrer restrictions
3. **Centralized Management**: Admin controls the key in one place
4. **Monitoring**: All usage tracked under one key
5. **Cost Control**: Usage quotas prevent abuse

## Deployment Checklist

### Google Cloud Console Setup

1. ✅ Enable Maps JavaScript API
2. ✅ Enable Places API
3. ✅ Create API Key
4. ✅ Set HTTP referrer restrictions:
   - Add production domain: `https://yourdomain.com/*`
   - Add development domain (if needed): `http://localhost:3001/*`
5. ✅ Set API restrictions: "Maps JavaScript API" + "Places API"
6. ✅ Set usage quotas
7. ✅ Set up billing alerts

### Backend Setup

1. ✅ Add `GOOGLE_MAPS_API_KEY` to `.env` file
2. ✅ Deploy backend with updated code
3. ✅ Verify `/api/maps/config` endpoint returns key

### Frontend Setup

1. ✅ Deploy frontend with updated code
2. ✅ Test maps functionality
3. ✅ Verify key is fetched from backend

## Monitoring

### Regular Checks

1. **Usage Monitoring**: Check Google Cloud Console dashboard
2. **Cost Monitoring**: Review billing reports monthly
3. **Error Monitoring**: Check for API key errors in logs
4. **Referrer Logs**: Review blocked referrers for suspicious activity

### Alerts to Set Up

1. Billing alerts at 50%, 75%, 90% of budget
2. Usage alerts at 80% of quota
3. Error rate alerts for Maps API

## Rollback Plan

If issues arise, you can temporarily revert by:

1. Frontend: Uncomment API key modals in components
2. Frontend: Change `useEffect` hooks back to `apiKeyService.get()`
3. Users can then provide their own keys again

## Future Enhancements

1. **Rate Limiting**: Add per-user rate limiting on backend
2. **Caching**: Cache geocoding results to reduce API calls
3. **Analytics**: Track which features use most API calls
4. **Alternative Providers**: Consider fallback to other map providers

## Related Files

### Backend
- `triply-server/.env` (API key storage)
- `triply-server/internal/config/config.go` (Configuration)
- `triply-server/internal/handlers/maps_handler.go` (Handler)
- `triply-server/cmd/server/main.go` (Route registration)

### Frontend
- `triply/lib/services/apiKeyService.ts` (Service)
- `triply/components/TripEditor.tsx` (Consumer)
- `triply/app/explore/[tripId]/page.tsx` (Consumer)

### Documentation
- `triply/README.md` (User guide)
- `triply-server/README.md` (Admin guide)

