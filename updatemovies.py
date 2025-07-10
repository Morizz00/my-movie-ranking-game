import csv
import requests
import urllib.parse
import time

OMDB_API_KEY = "3c26ba36"
FALLBACK_URL = "https://via.placeholder.com/300x450?text=No+Image"

def get_poster_url(title, content_type=None):
    try:
        encoded_title = urllib.parse.quote(title)
        
       
        time.sleep(0.2)
        
       
        if content_type and content_type.lower() in ["tv", "series", "tv show"]:
            api_type = "series"
        elif content_type and content_type.lower() in ["movie", "film"]:
            api_type = "movie"
        else:
            api_type = None
        
      
        if api_type:
            print(f"  Searching as {api_type}: {title}")
            response = requests.get(
                f"http://www.omdbapi.com/?t={encoded_title}&type={api_type}&apikey={OMDB_API_KEY}"
            )
            data = response.json()
            
            if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
                print(f"  ✓ Found poster for {title}")
                return data["Poster"]
            else:
                print(f"  ✗ No poster found as {api_type}: {data.get('Error', 'Unknown error')}")
        
        
        print(f"  Trying general search: {title}")
        response = requests.get(
            f"http://www.omdbapi.com/?t={encoded_title}&apikey={OMDB_API_KEY}"
        )
        data = response.json()
        
        if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
            print(f"  ✓ Found poster in general search for {title}")
            return data["Poster"]
        
       
        for search_type in ["movie", "series"]:
            print(f"  Trying as {search_type}: {title}")
            response = requests.get(
                f"http://www.omdbapi.com/?t={encoded_title}&type={search_type}&apikey={OMDB_API_KEY}"
            )
            data = response.json()
            
            if data.get("Response") == "True" and data.get("Poster") and data["Poster"] != "N/A":
                print(f"  ✓ Found poster as {search_type} for {title}")
                return data["Poster"]
        
        print(f"  ✗ No poster found for {title}")
        return FALLBACK_URL
        
    except Exception as e:
        print(f"  ✗ Error fetching poster for {title}: {e}")
        return FALLBACK_URL

def update_csv_file(input_file, output_file):
    print(f"Reading from {input_file}...")
    
    with open(input_file, "r", encoding="utf-8") as f:
        reader = csv.reader(f)
        headers = next(reader)
        rows = list(reader)
    
    print(f"Found {len(rows)} items to process")
    
    for i, row in enumerate(rows, 1):
        if len(row) < 4:
            print(f"Skipping row {i}: insufficient columns")
            continue
            
        title = row[1]
        print(f"[{i}/{len(rows)}] Processing: {title}")
        
        
        content_type = None
        if len(row) > 4:
            content_type = row[4]
        
        
        poster_url = get_poster_url(title, content_type)
        row[3] = poster_url
        
        
        if i % 10 == 0:
            print(f"Processed {i}/{len(rows)} items...")
    
    print(f"Saving to {output_file}...")
    with open(output_file, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(headers)
        writer.writerows(rows)
    
    print(f"✓ Updated file saved as {output_file}")


if __name__ == "__main__":
    try:
        update_csv_file("movies.csv", "movies_updated.csv")
    except FileNotFoundError:
        print("movies.csv not found, skipping...")
    
    # Update TV shows
    try:
        update_csv_file("tvshows.csv", "tvshows_updated.csv")
    except FileNotFoundError:
        print("tvshows.csv not found, skipping...")
    
    print("All updates completed!")
