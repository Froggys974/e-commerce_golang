-- produits de demo relancable sans doublon
INSERT INTO products (ref, name, description, category, price_ht_cents, vat_percent) VALUES
    ('PDT-KBRD27', 'Clavier mécanique', 'Clavier mécanique AZERTY, switches brown, rétroéclairé', 'informatique', 7499, 20),
    ('PDT-MSE384', 'Souris sans fil', 'Souris ergonomique sans fil, 6 boutons, autonomie 70 jours', 'informatique', 3332, 20),
    ('PDT-SCRN27', 'Écran 27 pouces', 'Écran IPS 27 pouces QHD 144 Hz', 'informatique', 24999, 20),
    ('PDT-HDST55', 'Casque audio', 'Casque sans fil à réduction de bruit active', 'informatique', 6666, 20),
    ('PDT-WBCM22', 'Webcam HD', 'Webcam 1080p avec micro intégré', 'informatique', 4999, 20),
    ('PDT-SSD2TB', 'SSD 2 To', 'SSD NVMe 2 To, lecture 7000 Mo/s', 'informatique', 12499, 20),
    ('PDT-USBC36', 'Hub USB-C', 'Hub USB-C 7 en 1 : HDMI, USB 3, lecteur SD', 'informatique', 2499, 20),
    ('PDT-BKGXLA', 'Le langage Go', 'Livre pour apprendre Go, des bases à la concurrence', 'livres', 3790, 5),
    ('PDT-BKCLNC', 'Clean Code', 'Livre de référence sur le code propre et lisible', 'livres', 3315, 5),
    ('PDT-BKDDDX', 'Domain-Driven Design', 'Livre sur la conception logicielle orientée métier', 'livres', 4740, 5),
    ('PDT-CAFE25', 'Café en grains', 'Café arabica en grains, 1 kg, torréfaction artisanale', 'alimentation', 1137, 5),
    ('PDT-THEVRT', 'Thé vert', 'Thé vert sencha bio, boîte de 100 g', 'alimentation', 853, 5),
    ('PDT-TSHG42', 'T-shirt Gopher', 'T-shirt en coton bio avec la mascotte Go', 'vetements', 1658, 20),
    ('PDT-HDYG48', 'Sweat à capuche', 'Sweat à capuche brodé, coton bio', 'vetements', 3749, 20),
    ('PDT-JDSC34', 'Jeu de société', 'Jeu de stratégie de 2 à 4 joueurs, parties de 45 minutes', 'jeux', 2915, 20)
ON CONFLICT (ref) DO NOTHING;
