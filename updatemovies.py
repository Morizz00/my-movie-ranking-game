import csv
import requests
import urllib.parse

OMDB_API_KEY = "3c26ba36"
FALLBACK_URL = "https://via.placeholder.com/300x450?text=No+Image"

def get_poster_url(title, content_type=None):
    try:
        encoded_title = urllib.parse.quote(title)
        
        response = requests.get(
            f"http://www.omdbapi.com/?t={encoded_title}&apikey={OMDB_API_KEY}"
        )
        data = response.json()
        
       
        if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
            return data["Poster"]
        
       
        if data.get("Response") == "False" and not content_type:
            
            response = requests.get(
                f"http://www.omdbapi.com/?t={encoded_title}&type=movie&apikey={OMDB_API_KEY}"
            )
            data = response.json()
            
            if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
                return data["Poster"]
            
            
            response = requests.get(
                f"http://www.omdbapi.com/?t={encoded_title}&type=series&apikey={OMDB_API_KEY}"
            )
            data = response.json()
            
            if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
                return data["Poster"]
        
    
        elif content_type:
            api_type = "series" if content_type.lower() in ["tv", "series", "tv show"] else "movie"
            response = requests.get(
                f"http://www.omdbapi.com/?t={encoded_title}&type={api_type}&apikey={OMDB_API_KEY}"
            )
            data = response.json()
            
            if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
                return data["Poster"]
        
        return FALLBACK_URL
        
    except Exception as e:
        print(f"Error fetching poster for {title}: {e}")
        return FALLBACK_URL


with open("movies.csv", "r", encoding="utf-8") as f:
    reader = csv.reader(f)
    headers = next(reader)
    movies = list(reader)


for movie in movies:
    title = movie[1]
    
    
    content_type = None
    if len(movie) > 4:  
        content_type = movie[4]
    
    print(f"Fetching poster for {title}...")
    movie[3] = get_poster_url(title, content_type)


with open("movies_updated.csv", "w", newline="", encoding="utf-8") as f:
    writer = csv.writer(f)
    writer.writerow(headers)
    writer.writerows(movies)

print("Updated movies.csv saved as movies_updated.csv")
