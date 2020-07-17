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
//   File:          GenerateHashes.java
//   Description:   Sample Java Source for GenerateHashes program
//
//  History:   
//    2016.09.13   adrian chandley  
//      Created java sample. 
//
/////////////////////////////////////////////////////////////////////////////////
 
import javax.imageio.*;
import java.awt.*;
import java.awt.image.*;
import java.io.*;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Arrays;
import java.util.Scanner;
import java.io.FileNotFoundException;

public class GenerateHashes {   

  public static final int  HASH_SIZE         = 144;
  public static final int  NCMEC_TARGET      = 41943;

  private static List<String> texttypes      = Arrays.asList("txt","lst","text","list");
  private static List<String> imagetypes     = Arrays.asList("jpg","bmp","gif","png");

  public static void main(String[] args) 
  {
    if (args.length == 0)
    {
                System.out.println("");
                System.out.println("Usage: ");
                System.out.println("");
                System.out.println("   java GenerateHashes <item> [<item> <..>]");
                System.out.println("");
                System.out.println("Where item may be:");
                System.out.println("    <image file>");
                System.out.println("    <directory of images>");
                System.out.println("    <txt file containing a list of image filenames>");
                System.out.println("");
    }
    else
    {
      PhotoDNA pdna = new PhotoDNA(4000*3000);
//      PhotoDNA pdna = new PhotoDNA();
      
      for (String arg : args) 
      {
        ScanLine(arg, pdna); 
      } 
    }
  }

  private static void ScanLine(String line, PhotoDNA pdna)
  {
    String test = line.trim();
    if (test.length() > 0) 
    {
      File f = new File(test);
      if (f.exists())
      {
        if (f.isDirectory()) {                   // directory
          File[] directoryListing = f.listFiles();
          if (directoryListing != null) {
            for (File child : directoryListing) {
              ScanLine(child.getPath(), pdna);
            }
          }
        }
        else {
          String filetype = getFileExtension(f);
          if (texttypes.contains(filetype))             // text file
          {
            try
            {
              Scanner txtfile = new Scanner(new File(test));
              while(txtfile.hasNextLine())
              {
                String ln = txtfile.nextLine();
                ScanLine(ln, pdna);
              }
              txtfile.close();
            }
            catch (IOException e) 
            {
              System.err.println("File "+test+": "+e.getMessage());
            }
          }
          else if (imagetypes.contains(filetype)) {
            byte[] hash = new byte[144];
            int res=new GenerateHashes().CalculateHash(f, pdna, hash);

            if (res == 0) DisplayHash(test,hash);
	     else System.err.println("Error calculating hash: "+res);
          } 
          else System.err.println("Unrecognised type: "+test);
        }
      }
      else System.err.println("Not found: "+test);
    }
  } 

  private static String getFileExtension(File file) {
    String name = file.getName();
    try {
        return name.substring(name.lastIndexOf(".") + 1).toLowerCase();
    } catch (Exception e) {
        return "";
    }
  }

  private static void DisplayHash(String file, byte[] hash)
  {
    System.out.print(file);
    for (int i = 0 ; i < HASH_SIZE; i++)  System.out.print("," + (hash[i] & 0xff));
    System.out.println("");
  }

  private int CalculateHash(String file, PhotoDNA pdna, byte[] hash)
  {
     File f = new File(file);
     return CalculateHash(f, pdna, hash);
  }

  private int CalculateHash(File f, PhotoDNA pdna, byte[] hash)
  {
    BufferedImage img = null;

    try {
	img = ImageIO.read(f);
    } catch (IOException e) {
	System.out.println("open failed");
      return -1;
    }  

    int w = img.getWidth(null);
    int h = img.getHeight(null);
    int tw = w;
    int th = h;
    int res = 0;

    if (img.getType() == BufferedImage.TYPE_3BYTE_BGR) {
      // get DataBufferBytes from Raster
      byte[] data = ((DataBufferByte) img.getRaster().getDataBuffer()).getData();
      res=pdna.ComputeHash(data, tw, th, 0, hash);
    } else {
      BufferedImage bgrImg = new BufferedImage(w, h, BufferedImage.TYPE_3BYTE_BGR);
      Graphics g = bgrImg.getGraphics();  
      g.drawImage(img, 0, 0, null);  
      g.dispose();  
      byte[] data = ((DataBufferByte) bgrImg.getRaster().getDataBuffer()).getData();
      res=pdna.ComputeHash(data, tw, th, 0, hash);
    }
    img = null;
    return res;
  }

}
