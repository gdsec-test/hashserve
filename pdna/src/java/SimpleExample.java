/////////////////////////////////////////////////////////////////////////////////
// 
// IMPORTANT NOTICE
// ================
// 
//   Copyright (C) 2016, 2017, Microsoft Corporation
//   All Rights Reserved.
// 
//   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS STRICTLY UNDER THE TERMS
//   OF A SEPARATE LICENSE AGREEMENT.  NO SOURCE CODE DISTRIBUTION OR SOURCE 
//   CODE DISCLOSURE RIGHTS ARE PROVIDED WITH RESPECT TO THIS SOFTWARE OR ANY 
//   DERIVATIVE WORKS THEREOF.  USE AND NON-SOURCE CODE REDISTRIBUTION OF THE 
//   SOFTWARE IS PERMITTED ONLY UNDER THE TERMS OF SUCH LICENSE AGREEMENT.  
//   THIS SOFTWARE AND ANY WORKS/MODIFICATIONS/TRANSLATIONS DERIVED FROM THIS 
//   SOFTWARE ARE COVERED BY SUCH LICENSE AGREEMENT AND MUST ALSO CONTAIN THIS 
//   NOTICE.  THIS SOFTWARE IS CONFIDENTIAL AND SENSITIVE AND MUST BE SECURED 
//   WITH LIMITED ACCESS PURSUANT TO THE TERMS OF SUCH LICENSE AGREEMENT.
// 
//   THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
//   INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY 
//   AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
//   COPYRIGHT HOLDER BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, 
//   EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, 
//   PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
//   OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
//   WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR 
//   OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF 
//   ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//   
//   Project:       PhotoDNA Robust Hashing
//   File:          SimpleExample.java
//   Description:   Simple Java Source sample to compare two images
//
//  History:   
//    2016.09.13   adrian chandley
//      Created simple Java example.
//
/////////////////////////////////////////////////////////////////////////////////
 
import javax.imageio.*;
import java.awt.*;
import java.awt.image.*;
import java.io.*;
import java.io.FileNotFoundException;

public class SimpleExample {   

  public static final int HASH_SIZE = 144;
  
  public static void main(String[] args) 
  {
    File infile;
    
// initialse photoDNA
    PhotoDNA rh = new PhotoDNA();

// initialise buffers
    byte[] hash1 = new byte[HASH_SIZE];
    byte[] hash2 = new byte[HASH_SIZE];
    BufferedImage img = null;

// read in image
    infile = new File("tmp/Tests/Test01.jpg");
    try {
      img = ImageIO.read(infile);
    } catch (IOException e) {
      return;
    }  
    
// Calculate and display the hash
    int res = rh.ComputeHash(((DataBufferByte) img.getRaster().getDataBuffer()).getData(), 
                img.getWidth(null), img.getHeight(null), 0, hash1);
    DisplayHash("Test01", hash1);
   
// read in image
    infile = new File("tmp/Tests/Test02.jpg");
    try {
      img = ImageIO.read(infile);
    } catch (IOException e) {
      return;
    }  

// Calculate and display the hash
    res = rh.ComputeHash(((DataBufferByte) img.getRaster().getDataBuffer()).getData(), 
                img.getWidth(null), img.getHeight(null), 0, hash2);
    DisplayHash("Test02", hash2);
  
// Display the distance
    System.out.println("Distance between hashes: " + Distance(hash1,hash2)) ;
  }
  
  private static void DisplayHash(String name, byte[] hash)
  {
    System.out.print(name);
    for (int i = 0 ; i < HASH_SIZE; i++)  System.out.print("," + (hash[i] & 0xff));
    System.out.println("\n");
  }

  private static int Distance(byte[] hash1, byte[] hash2)
  {
    int distance = 0;
    for (int i = 0; i < HASH_SIZE; i++) {
      int diff = (0xFF & hash1[i]) - (0xFF & hash2[i]);
      distance += diff * diff;
    }
    return distance;
  }

}
