import type { Flight, ConnectingFlight, Airport, City } from './client';

interface GqlResponse<T> {
  data: T;
  errors?: { message: string }[];
}

const GQL_BASE = import.meta.env.VITE_API_URL || '';

async function gqlRequest<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const res = await fetch(`${GQL_BASE}/graphql`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, variables }),
  });
  const json: GqlResponse<T> = await res.json();
  if (json.errors?.length) {
    throw new Error(json.errors[0].message);
  }
  return json.data;
}

export interface SearchFlightsResult {
  searchFlights: {
    flights: Flight[];
    connecting: ConnectingFlight[];
    cached: boolean;
  };
}

export interface AirportsResult {
  airports: Airport[];
}

export interface CitiesResult {
  cities: City[];
}

export interface AirportsAndCitiesResult {
  airports: Airport[];
  cities: City[];
}

const SEARCH_FLIGHTS_QUERY = `
  query SearchFlights($origin_id: String!, $destination_id: String!, $date: String!) {
    searchFlights(origin_id: $origin_id, destination_id: $destination_id, date: $date) {
      flights {
        id flight_number origin_id destination_id departure_time arrival_time price seats_total seats_available
      }
      connecting {
        leg1 { id flight_number origin_id destination_id departure_time arrival_time price seats_total seats_available }
        leg2 { id flight_number origin_id destination_id departure_time arrival_time price seats_total seats_available }
        total_price discount total_duration_hours layover_minutes layover_airport_id
      }
      cached
    }
  }
`;

const AIRPORTS_AND_CITIES_QUERY = `
  query AirportsAndCities {
    airports { id name city_id code }
    cities { id name country code }
  }
`;

export const gqlApi = {
  searchFlights: (originId: string, destId: string, date: string) =>
    gqlRequest<SearchFlightsResult>(SEARCH_FLIGHTS_QUERY, {
      origin_id: originId,
      destination_id: destId,
      date,
    }),

  airportsAndCities: () =>
    gqlRequest<AirportsAndCitiesResult>(AIRPORTS_AND_CITIES_QUERY),
};
